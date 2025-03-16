from models import LogEvents

def get_latest_logs(cloudwatch_client, log_group_name: str, n: int, grepv: str = None, grep: str = None) -> LogEvents:
    # Get the list of log streams in the log group, sorted by last event time
    response = cloudwatch_client.describe_log_streams(
        logGroupName=log_group_name,
        orderBy='LastEventTime',
        descending=True,
        limit=50  # Increase limit to get more log streams
    )
    
    log_streams = response.get('logStreams', [])
    if not log_streams:
        return []

    all_events = []
    for log_stream in log_streams:
        if len(all_events) >= n:
            break

        # Get the log events from the current log stream
        response = cloudwatch_client.get_log_events(
            logGroupName=log_group_name,
            logStreamName=log_stream['logStreamName'],
            limit=n - len(all_events),
            startFromHead=False
        )

        validated = LogEvents.model_validate(response, strict=False)

        if grepv:
            for event in validated.events:
                if grepv not in event.message:
                    all_events.append(event)
        else:
            all_events.extend(validated.events)

    if grep:
        all_events = [event for event in all_events if grep in event.message]

    return all_events[:n]

