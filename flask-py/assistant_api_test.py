import openai
import time


def create_thread(client):
    """
    Creates a new thread using the provided OpenAI client.

    Args:
        client: The OpenAI client instance to use for creating the thread.

    Returns:
        The created thread object.
    """
    my_thread = client.beta.threads.create()
    return my_thread



def wait_for_run_completion(client, thread_id, poll_interval=0.1, timeout=5):
    """
    Waits for the latest run in a thread to either complete or fail, with a timeout.

    Args:
        client: The OpenAI client instance.
        thread_id: The ID of the thread to monitor.
        poll_interval: Time in seconds to wait between each status check.
        timeout: Maximum time to wait for run completion in seconds.

    Returns:
        A tuple containing a boolean indicating success and an error message if any.
    """
    start_time = time.time()

    while True:
        # Check if the timeout is exceeded
        if time.time() - start_time > timeout:
            return False, "Timeout exceeded while waiting for run to complete."

        # Check the run status
        runs = client.beta.threads.runs.list(thread_id)
        latest_run = runs.data[0]
        if latest_run.status == "completed":
            return True, None  # Successful completion
        elif latest_run.status == "failed":
            return False, "Run failed."

        time.sleep(poll_interval)  # Wait before the next check



def create_message_and_run_thread(client, thread_id, assistant_id, content):
    """
    Creates a message in a thread, initiates a run, waits for completion,
    and retrieves the latest messages in the thread.

    Args:
        client: The OpenAI client instance.
        thread_id: The ID of the thread to interact with.
        assistant_id: The ID of the assistant to run.
        content: The content of the message to be sent.

    Returns:
        The latest message text or None if no message is retrieved or if an error occurs.
    """

    # Send the user's message to the thread
    start = time.time()
    client.beta.threads.messages.create(thread_id=thread_id, role="user", content=content)
    print(f"Time to create message: {time.time() - start} seconds")

    # Initiate a run with the assistant
    start = time.time()
    client.beta.threads.runs.create(thread_id=thread_id, assistant_id=assistant_id)
    print(f"Time to create run: {time.time() - start} seconds")

    # Wait for the run to complete or fail
    start = time.time()
    success, error = wait_for_run_completion(client, thread_id)
    if not success:
        print(f"Error: {error}")
        return None
    print(f"Time to wait for run completion: {time.time() - start} seconds")

    # Retrieve and process the messages from the thread
    start = time.time()
    response = client.beta.threads.messages.list(thread_id=thread_id)
    print(f"Time to retrieve messages: {time.time() - start} seconds")

    # Iterate through the messages and return the latest message text
    for message in response.data:  # Reverse the order to get the latest message first
        message_text = message.content[0].text.value
        if message_text.strip() != "":
            return message_text.strip()

    return None


def main():
    start_time = time.time()

    # Set your OpenAI API key and organization ID
    openai.organization = ''

    # Specify the assistant ID
    assistant_id = "asst_xIHAFLR0eWlTYrRcIeoG0xvj"

    # Initialize the OpenAI client
    client = openai.Client(api_key="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")

    # Retrieve details of the assistant
    start = time.time()
    my_assistant = client.beta.assistants.retrieve(assistant_id)
    end = time.time()
    print("Assistant Details:", my_assistant)
    print(f"Retrieving assistant details took {end - start} seconds")

    # Create thread
    start = time.time()
    my_thread = create_thread(client)
    end = time.time()
    print("Thread Created:", my_thread)
    print(f"Creating thread took {end - start} seconds")

    # Interaction 1
    content = "what is your name?"
    start = time.time()
    res =  create_message_and_run_thread(client, my_thread.id, assistant_id, content)
    print(res)
    end = time.time()
    print(f"Interaction with content '{content}' took {end - start} seconds")

    # Interaction 2
    content = "what can yo do?"
    start = time.time()
    res =  create_message_and_run_thread(client, my_thread.id, assistant_id, content)
    print(res)
    end = time.time()
    print(f"Interaction with content '{content}' took {end - start} seconds")

    # Interaction 3
    content = "what is my first question?"
    start = time.time()
    res =  create_message_and_run_thread(client, my_thread.id, assistant_id, content)
    print(res)
    end = time.time()
    print(f"Interaction with content '{content}' took {end - start} seconds")

    total_end_time = time.time()
    print(f"Total execution time: {total_end_time - start_time} seconds")

# Call the main function
main()