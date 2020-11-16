FROM python:3

ADD eventslog/server.py /

RUN pip install pynsq

CMD [ "python", "./server.py" ]