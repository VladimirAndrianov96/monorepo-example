import nsq

buf = []

def process_message(message):
    global buf
    message.enable_async()
    # cache the message for later processing
    buf.append(message)
    if len(buf) >= 3:
        for msg in buf:
            print (msg)
            msg.finish()
        buf = []
    else:
        print ('deferring processing')

r = nsq.Reader(message_handler=process_message,
        lookupd_http_addresses=['http://nsqlookupd:4161'],
        topic='events', channel="basic_logs_python", max_in_flight=9)
nsq.run()