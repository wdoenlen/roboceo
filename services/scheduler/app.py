import os
import random

import asana
from flask import Flask, request

app = Flask(__name__)

CEO_ACCESS_TOKEN = os.environ.get("CEO_ASANA_ACCESS_TOKEN")
WORKSPACE_ID = int(os.environ.get("EXECUTIVE_MACHINE_ASANA_WORKSPACE_ID"))

def get_tasks_with_tag(client, context, is_work):

    tag_ids = {}
    for entry in client.tags.find_all({"workspace": WORKSPACE_ID}):
        name = entry['name'].lower()
        tag_ids[name] = entry['id']

    if context not in tag_ids:
        return []

    if not is_work:
        context = "not work"

    tag_id = tag_ids[context]

    names = [task["name"] for task in client.tags.get_tasks_with_tag(tag_id)]

    return names

@app.route("/task")
def get_task():
    context = request.args["context"].lower()
    is_work = False if request.args.get("work", True) == "false" else True

    client = asana.Client.access_token(CEO_ACCESS_TOKEN)
    tasks = get_tasks_with_tag(client, context, is_work)
    if len(tasks) == 0 and context != 'anywhere':
        tasks = get_tasks_with_tag('anywhere')
    if len(tasks) == 0:
        tasks = ["Brainstorm next steps"]

    next_task = random.choice(tasks)

    return next_task

if __name__ == "__main__":
    app.run(host='0.0.0.0')
