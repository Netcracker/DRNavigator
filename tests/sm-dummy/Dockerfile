FROM --platform=$BUILDPLATFORM python:3.13.5-alpine3.22

USER root

ARG USER_NAME=smdumme
ARG USER_UID=1001

COPY sm-dummy /app/
COPY requirements.txt /app/requirements.txt

WORKDIR /app/

RUN apk update && \
    apk upgrade && \
    apk add bash && \
    pip install --upgrade pip==24.* --no-cache-dir && \
    pip install -r requirements.txt --no-cache-dir && \
    adduser -D -u "${USER_UID}" "${USER_NAME}"

USER ${USER_UID}

CMD ["python", "sm-dummy"]
