FROM dos-client:latest AS client

FROM python:3.12-slim

WORKDIR /work

COPY --from=client /usr/local/bin/dos /usr/local/bin/dos
COPY --from=client /work/config.yml /work/config.yml

RUN pip install --no-cache-dir pytest

COPY tests /work/tests

CMD ["pytest", "tests/e2e"]
