import requests
import os

def create_thread(base_url):
    """Create a thread using the Flask endpoint."""
    url = f"{base_url}/create_thread"
    response = requests.get(url)
    if response.status_code == 200:
        return response.json()
    else:
        print("Failed to create thread:", response.text)
        return None

def create_message_and_run_thread(base_url, thread_id, assistant_id, content):
    """Create a message and run a thread using the Flask endpoint."""
    url = f"{base_url}/create_message_and_run_thread"
    data = {
        "thread_id": thread_id,
        "assistant_id": assistant_id,
        "content": content
    }
    response = requests.post(url, json=data)
    if response.status_code == 200:
        return response.json()
    else:
        print("Failed to create message and run thread:", response.text)
        return None

def main():
    # Set the base URL for your Flask app
    base_url = "http://localhost:8085"  # Update with your Flask app URL and port

    # Test creating a thread
    thread_response = create_thread(base_url)
    if thread_response:
        print("Thread created successfully:", thread_response)
        thread_id = thread_response['thread_id']

        # Test creating a message and running a thread
        assistant_id = "asst_xIHAFLR0eWlTYrRcIeoG0xvj"
        content = "Hello, what's your name?"
        message_response = create_message_and_run_thread(base_url, thread_id, assistant_id, content)
        if message_response:
            print("Message created and thread run successfully:", message_response)

        content = "what is my first question?"
        message_response = create_message_and_run_thread(base_url, thread_id, assistant_id, content)
        if message_response:
            print("Message created and thread run successfully:", message_response)


if __name__ == '__main__':
    main()