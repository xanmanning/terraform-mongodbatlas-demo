{
    "connection_strings": {%{ for key, collection in service_config.mongoCollection }
        "${service_config.mongoDatabase}.${collection}": "mongodb+srv://${service_config.serviceName}:${password}@${connection_string}/${service_config.mongoDatabase}/${collection}"%{ if key >= 0 && (key + 1) < length(service_config.mongoCollection) },%{ endif }%{ endfor }
    }
}
