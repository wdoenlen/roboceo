#!/usr/bin/env python

"""Searches for tasks in Asana that the CEO is not following and automatically
adds the CEO as a follower on the task.

This could probably be done better with something like a webhook after a task
has been created.
"""

import asana
import constants
import common

def autofollow():

    client = common.build_asana_client()
    projects = client.projects.find_all({"workspace": constants.WORKSPACE_ID})
    task_list = []

    for project in projects:
        tasks = list(client.tasks.find_all({"project": project["id"]}))
        task_list += tasks

    for task in task_list:
        client.tasks.add_followers(task["id"], {"followers":
                                                [constants.CEO_USER_ID]})

if __name__ == "__main__":
    autofollow()
