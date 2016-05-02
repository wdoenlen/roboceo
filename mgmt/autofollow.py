#!/usr/bin/env python

"""Searches for tasks in Asana that the CEO is not following and automatically
adds the CEO as a follower on the task.

This could probably be done better with something like a webhook after a task
has been created.
"""

import asana
import constants
import common
import datetime
from pytz import timezone

time_interval = 1 # day

def autofollow():

    client = common.build_asana_client()
    projects = client.projects.find_all({"workspace": constants.WORKSPACE_ID})
    task_list = []

    modified_since_dt = (datetime.datetime.now(timezone("US/Pacific")) -
                            datetime.timedelta(days=time_interval))
    modified_since_dt = modified_since_dt.replace(microsecond=0)
    modified_since_str = modified_since_dt.isoformat()

    # We can't get a list of all tasks in a workspace without either specifying
    # the projects the tasks are in or who the tasks are assigned to.
    for project in projects:
        data = {"project": project["id"], "modified_since": modified_since_str}
        tasks = list(client.tasks.find_all(data))
        task_list += tasks

    if task_list:
        for task in task_list:
            client.tasks.add_followers(task["id"], {"followers":
                                                    [constants.CEO_USER_ID]})

if __name__ == "__main__":
    autofollow()
