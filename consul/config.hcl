consul {
  retry {
    enabled = true
    attempts = 0
    backoff = "1s"
    max_backoff = "1m"
  }

  ssl {
    enabled = true
  }
}

wait {
  min = "10s"
}

template {
  source = "consul/utm_postback_middle.yml.ctmpl"
  destination = "configs/utm_postback_middle.yml"
  error_on_missing_key = true
}