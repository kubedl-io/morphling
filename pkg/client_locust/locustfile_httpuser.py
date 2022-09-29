import os
from locust import HttpUser, task

default_host = os.environ["ServiceName"]
if not default_host.startswith('http://'):
    default_host = 'http://' + default_host

class MyHttpUser(HttpUser):
    host = os.getenv("HTTP_HOST", default_host)
    assert host != ""

    @task
    def access(self):
        self.client.get("/")