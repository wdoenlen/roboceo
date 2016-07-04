"""
Every so often, this will

[{"start", "duration", "activity"}]
- Existing log of what's happened (pickled something or other)
"""

import datetime
import pickle
import random
import time

from activities import business_activities, engineering_activities, networking_activities

import asana
import constants

def build_asana_client():
    """Connects to Asana and returns an API client"""
    return asana.Client.access_token(constants.CEO_ACCESS_TOKEN)

def get_tasks(context):
    pass

def get_current_context():
    pass

import os

CEO_ACCESS_TOKEN = os.environ.get("CEO_ASANA_ACCESS_TOKEN")
WORKSPACE_ID = int(os.environ.get("EXECUTIVE_MACHINE_ASANA_WORKSPACE_ID"))
CEO_USER_ID = int(os.environ.get("CEO_ASANA_USER_ID"))
WILL_USER_ID = int(os.environ.get("WILL_ASANA_USER_ID"))
MAX_USER_ID = int(os.environ.get("MAX_ASANA_USER_ID"))
MAX_EMAIL = os.environ.get("MAX_EMAIL")
WILL_EMAIL = os.environ.get("WILL_EMAIL")

tags = {
    "Internet": 151174408734357
}

if __name__ == "__main__":

    # with open("db.pickle", "r") as pickle_file:
    #     previous_activity = pickle.load(pickle_file)

    client = build_asana_client()

    # Check the context we're currently in
    context = get_current_context()

    # Using the context get all tasks not in someday/maybe with that context
    tasks = get_tasks(context)

    # Pick a task
    next_task = random.choice(tasks)

    # Get task information
    title, description = get_task_info()

    return json.dumps({"title": title, "description": description})












    start_time = datetime.datetime.now()

    while True:
        # Choose whether or not we're doing business, engineering or networking
        # Look at current mix of biz/eng/net
        activities = {
            "business": business_activities,
            "networking": networking_activities,
            "engineering": engineering_activities,
        }
        next_activity_category = random.choice(activities.keys())
        next_activity = random.choice(activities[next_activity_category])
        next_duration = random.choice(range(15, 120))


        next_entry = {
            "start": start_time.isoformat(),
            "duration": next_duration,
            "activity_category": next_activity_category,
            "activity": next_activity,
        }

        print next_entry

        start_time += datetime.timedelta(minutes=next_duration)

        time.sleep(3)
