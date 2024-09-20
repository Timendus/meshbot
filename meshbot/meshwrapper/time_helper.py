from datetime import datetime
import math


def time_ago(timestamp):
    now = datetime.now()
    if type(timestamp) != datetime:
        timestamp = datetime.fromtimestamp(timestamp)
    seconds = math.floor((now - timestamp).total_seconds())
    if seconds == 1:
        return f"one second"
    if seconds < 60:
        return f"{str(seconds)} seconds"

    minutes = math.floor(seconds / 60)
    if minutes == 1:
        return f"one minute"
    if minutes < 60:
        return f"{str(minutes)} minutes"

    hours = math.floor(minutes / 60)
    if hours == 1:
        return f"one hour"
    if hours < 24:
        return f"{str(hours)} hours"

    days = math.floor(hours / 24)
    if days == 1:
        return f"one day"
    return f"{str(days)} days"