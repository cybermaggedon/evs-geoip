
FROM fedora:32

ARG EVS_API=master

RUN dnf install -y python3 python3-pip git && \
    dnf install -y python3-geoip2 geoipupdate && \
    dnf clean all && \
    pip3 install git+https://github.com/cybermaggedon/evs-python-api@${EVS_API}

COPY evs-geoip /usr/local/bin/

ENV PULSAR_BROKER=pulsar://exchange
ENV METRICS_PORT=8088
EXPOSE 8088

CMD /usr/local/bin/evs-geoip

