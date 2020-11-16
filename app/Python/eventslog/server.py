import nsq

def new_user_handler(message):
    print ("New: " + str(message.body))
    return True

def deactivated_user_handler(message):
    print ("Deactivated: " + str(message.body))
    return True

def activated_user_handler(message):
    print ("Activated: " + str(message.body))
    return True

r1 = nsq.Reader(message_handler=new_user_handler,
        lookupd_http_addresses=['nsqlookupd:4161'],
        topic='new_user', channel='events-python', lookupd_poll_interval=15)

r2 = nsq.Reader(message_handler=deactivated_user_handler,
        lookupd_http_addresses=['nsqlookupd:4161'],
        topic='deactivated_user', channel='deactivated_user', lookupd_poll_interval=15)

r3 = nsq.Reader(message_handler=activated_user_handler,
        lookupd_http_addresses=['nsqlookupd:4161'],
        topic='activated_user', channel='events-python', lookupd_poll_interval=15)

nsq.run()