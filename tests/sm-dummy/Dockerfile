FROM python:3.8-slim

USER root

COPY tests/sm-dummy/sm-dummy /app/
COPY tests/sm-dummy/requirements.txt /app/requirements.txt

WORKDIR /app/

RUN pip3 install --upgrade pip && \
    pip3 install -r requirements.txt

CMD ["python", "sm-dummy"]
