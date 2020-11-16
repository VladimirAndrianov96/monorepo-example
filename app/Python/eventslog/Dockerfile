FROM python:3.9-alpine

ADD eventslog/server.py /

RUN pip install pynsq

CMD [ "python", "-u", "./server.py" ]