package utm_postback_middle

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/maximus335/utm_postback_middle/docs"
	"github.com/maximus335/utm_postback_middle/internal/app/utm_postback_middle/config"
	appsflyerdata "github.com/maximus335/utm_postback_middle/internal/pkg/appsflyer_data"
	"github.com/maximus335/utm_postback_middle/internal/pkg/consumer"
	"github.com/maximus335/utm_postback_middle/internal/pkg/db"
	"github.com/maximus335/utm_postback_middle/internal/pkg/events"
	"github.com/maximus335/utm_postback_middle/internal/pkg/types"
	"github.com/sirupsen/logrus"
	httpSwagger "github.com/swaggo/http-swagger"
)

const ctxKeyRequestID ctxKey = iota

type ctxKey int8

type UtmPostbackMiddle struct {
	config *config.Configuration
	logger *logrus.Logger
	router *mux.Router
}

func New(config *config.Configuration) *UtmPostbackMiddle {
	return &UtmPostbackMiddle{
		config: config,
		logger: logrus.New(),
		router: mux.NewRouter(),
	}
}

func (s *UtmPostbackMiddle) Start() error {
	s.configureLogger()

	db := s.openDBConnections()
	defer s.closeDBConnections(db)

	s.configureRoyter(db)

	s.StartKafka(db)

	s.logger.Info("Server Started")

	return http.ListenAndServe(s.config.Server.ListenAddr, s.router)
}

func (s *UtmPostbackMiddle) StartKafka(db *pgxpool.Pool) {
	consmr := consumer.New(&s.config.Kafka, &s.config.Appsflyer, s.logger, db)
	consmr.StartKafkaReaders()
	s.logger.Info("Kafka Started")
}

func (s *UtmPostbackMiddle) configureLogger() {
	s.logger.SetFormatter(&logrus.JSONFormatter{})
}

func (s *UtmPostbackMiddle) configureRoyter(db *pgxpool.Pool) {
	s.router.Use(s.setRequestID)
	s.router.Use(s.logRequest)
	api := s.router.PathPrefix("/api/v1").Subrouter()
	api.Use(s.commonMiddleware)
	swagger := s.router.PathPrefix("/swagger").Subrouter()
	swagger.Use(s.swaggerMiddleware)
	s.router.HandleFunc("/healthcheck", s.healthCheck()).Methods("Get")
	api.HandleFunc("/appsflyer", s.getAppsflyerData(db)).Methods("Get")
	api.HandleFunc("/appsflyer", s.postAppsflyerData(db)).Methods("Post")
	createEventHandler := http.HandlerFunc(s.createEvent(db))
	api.Handle("/event", s.authorize(createEventHandler)).Methods("Post")
	swagerHandlerFunc := httpSwagger.Handler(
		httpSwagger.URL(s.config.Server.Host+"/swagger/doc.json"), //The url pointing to API definition
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("#swagger-ui"),
	)
	swagger.PathPrefix("/").Handler(swagerHandlerFunc)
}

func (s *UtmPostbackMiddle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *UtmPostbackMiddle) authorize(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqToken := r.Header.Get("Auth-Token")
		if reqToken != s.config.Server.AuthToken {
			s.respond(w, r, http.StatusUnauthorized, nil)
		}
		next.ServeHTTP(w, r)
	})
}

func (s *UtmPostbackMiddle) healthCheck() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.respond(w, r, http.StatusOK, nil)
	}
}

func (s *UtmPostbackMiddle) getAppsflyerData(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		advertisingId := r.URL.Query().Get("advertising_id")
		androidId := r.URL.Query().Get("android_id")
		idfa := r.URL.Query().Get("idfa")
		dataType := r.URL.Query().Get("data_type")
		if dataType == "raw" {
			data, err := appsflyerdata.RawAppsflyerData(db, advertisingId, androidId, idfa)
			if err != nil {
				s.logger.Error(err)
				s.error(w, r, http.StatusInternalServerError, err)
				return
			}
			s.respond(w, r, http.StatusOK, data)
		} else {
			data, err := appsflyerdata.AllAppsflyerData(db, advertisingId, androidId, idfa)
			if err != nil {
				s.logger.Error(err)
				s.error(w, r, http.StatusInternalServerError, err)
				return
			}
			s.respond(w, r, http.StatusOK, data)
		}
	}
}

func (s *UtmPostbackMiddle) postAppsflyerData(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rawParams := s.extractJsonParams(w, r)

		if err := appsflyerdata.CreateAppsflyerData(db, rawParams); err != nil {
			s.logger.Error(err)
			s.error(w, r, http.StatusInternalServerError, err)
		} else {
			s.respond(w, r, http.StatusOK, nil)
		}
	}
}

func (s *UtmPostbackMiddle) createEvent(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rawParams := s.extractJsonParams(w, r)
		err := events.CreateEvent(db, rawParams, &s.config.Appsflyer)
		if err != nil {
			s.logger.Error(err)
			s.handleLogicError(w, r, err)
		} else {
			s.respond(w, r, http.StatusOK, nil)
		}
	}
}

func (s *UtmPostbackMiddle) handleLogicError(w http.ResponseWriter, r *http.Request, err error) {
	if _, ok := err.(*types.NotFoundError); ok {
		s.error(w, r, http.StatusNotFound, err)
	} else if _, ok := err.(*types.ValidationParamsError); ok {
		s.error(w, r, http.StatusBadRequest, err)
	} else {
		s.error(w, r, http.StatusInternalServerError, err)
	}
}

func (s *UtmPostbackMiddle) extractJsonParams(w http.ResponseWriter, r *http.Request) map[string]interface{} {
	var paramsMapTemplate interface{}

	if err := json.NewDecoder(r.Body).Decode(&paramsMapTemplate); err != nil {
		s.logger.Error(err)
		s.error(w, r, http.StatusBadRequest, err)
	}

	return paramsMapTemplate.(map[string]interface{})
}

func (s *UtmPostbackMiddle) closeDBConnections(db *pgxpool.Pool) {
	s.logger.Debug("Closing db connection")
	db.Close()
}

func (s *UtmPostbackMiddle) openDBConnections() *pgxpool.Pool {
	opts := NewDBConfig(s.config)
	db, err := db.Connect(opts)
	if err != nil {
		s.logger.Fatal(err)
	}

	return db
}

func (s *UtmPostbackMiddle) setRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := uuid.New().String()
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKeyRequestID, id)))
	})
}

func (s *UtmPostbackMiddle) swaggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		next.ServeHTTP(w, r)
	})
}

func (s *UtmPostbackMiddle) commonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func (s *UtmPostbackMiddle) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := s.logger.WithFields(logrus.Fields{
			"remote_addr": r.RemoteAddr,
			"request_id":  r.Context().Value(ctxKeyRequestID),
		})
		logger.Infof("started %s %s", r.Method, r.RequestURI)

		start := time.Now()
		rw := &responseWriter{w, http.StatusOK}
		next.ServeHTTP(rw, r)

		var level logrus.Level
		switch {
		case rw.code >= 500:
			level = logrus.ErrorLevel
		case rw.code >= 400:
			level = logrus.WarnLevel
		default:
			level = logrus.InfoLevel
		}
		logger.Logf(
			level,
			"completed with %d %s in %v",
			rw.code,
			http.StatusText(rw.code),
			time.Since(start),
		)
	})
}

func (s *UtmPostbackMiddle) error(w http.ResponseWriter, r *http.Request, code int, err error) {
	s.respond(w, r, code, map[string]string{"error": err.Error()})
}

func (s *UtmPostbackMiddle) respond(w http.ResponseWriter, r *http.Request, code int, data interface{}) {
	w.WriteHeader(code)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			s.logger.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func NewDBConfig(cfg *config.Configuration) *db.Options {
	return &db.Options{
		URL:             cfg.Database.Url,
		ConnMaxLifetime: time.Duration(cfg.Database.ConnMaxLifetime) * time.Second,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
	}
}
