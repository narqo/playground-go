#FROM aerospike:3.16.0.6
FROM aerospike:4.5.3.6

# default namespace that we're not interested in
ENV NAMESPACE=aerotest

RUN set -eu; \
        { \
# define "persistent" namespaces
            echo 'namespace persistent00 {'; \
            echo '	replication-factor 1'; \
            echo '	memory-size 20M'; \
            echo '	default-ttl 0'; \
            echo '	storage-engine memory'; \
            echo '}'; \
            echo; \
        } >> /etc/aerospike/aerospike.template.conf
