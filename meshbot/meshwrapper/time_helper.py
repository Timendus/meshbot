from datetime import datetime, timedelta
import math


def time_ago(timestamp):
    now = datetime.now()
    if timestamp == None:
        return "an unknown amount of time"
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


def friendly_date(date):
    today = datetime.now().date()
    if date.date() == today:
        return "Today"
    if date.date() == today + timedelta(days=1):
        return "Tomorrow"
    if date.date() < today + timedelta(days=7):
        return date.strftime("%a")
    return date.strftime("%d-%m-%Y")
