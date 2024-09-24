import sys
import os
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

import logging
from queries import Query
from scenarios import query_and_preserve

url = "http://192.168.188.42:8080"

# login train-ticket and store the cookies
q = Query(url)
if not q.login():
    logging.fatal('login failed')

# execute scenario on current user
query_and_preserve(q)

# or execute query directly
q.query_high_speed_ticket()