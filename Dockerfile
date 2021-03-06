
FROM fedora:32

RUN dnf install -y geoipupdate && \
    dnf clean all

COPY evs-geoip /usr/local/bin/

WORKDIR /usr/local/share/

ENV PULSAR_BROKER=pulsar://exchange:6650
ENV METRICS_PORT=8088

# MaxMinds licence key
ENV ACCOUNT_ID=0
ENV LICENCE_KEY=0
ENV EDITION_IDS="GeoLite2-ASN GeoLite2-City"

EXPOSE 8088

CMD ( \
      echo AccountID ${ACCOUNT_ID} && \
      echo LicenseKey ${LICENCE_KEY} && \
      echo EditionIDs ${EDITION_IDS} \
    ) > GeoIP.conf && \
    /usr/local/bin/evs-geoip

