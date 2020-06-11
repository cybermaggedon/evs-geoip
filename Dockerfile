
FROM fedora:32

ARG EVS_API=master

RUN dnf install -y geoipupdate && \
    dnf clean all

COPY evs-geoip /usr/local/bin/

COPY GeoLite2-City.mmdb /usr/local/share/
COPY GeoLite2-ASN.mmdb /usr/local/share/

WORKDIR /usr/local/share/

ENV PULSAR_BROKER=pulsar://exchange
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

