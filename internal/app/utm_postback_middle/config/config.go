package config

type Configuration struct {
	Server    ServerConfiguration
	Database  DatabaseConfiguration
	Appsflyer AppsflyerConfiguration
	Kafka     KafkaConfiguration
}

type DatabaseConfiguration struct {
	Url             string
	ConnMaxLifetime int
	MaxOpenConns    int32
}

type ServerConfiguration struct {
	ListenAddr string
	AuthToken  string
	Host       string
}

type AppsflyerConfiguration struct {
	Url    string
	ApiKey string
}

type KafkaConfiguration struct {
	Brokers []string
	Topics  []string
}
