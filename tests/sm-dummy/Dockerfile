FROM python:3.8-slim

USER root

ARG USER_NAME=smdumme
ARG USER_UID=1001

COPY sm-dummy /app/
COPY requirements.txt /app/requirements.txt

WORKDIR /app/

RUN pip3 install --upgrade pip && \
    pip3 install -r requirements.txt && \
    useradd -r -u "${USER_UID}" "${USER_NAME}"

USER ${USER_UID}

CMD ["python", "sm-dummy"]
