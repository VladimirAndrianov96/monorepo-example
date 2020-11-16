import nsq

def handler(message):
    print (message.body)
    return True

r = nsq.Reader(message_handler=handler,
        lookupd_http_addresses=['nsqlookupd:4161'],
        topic='activated_user', channel='events-python', lookupd_poll_interval=5)

nsq.run()


