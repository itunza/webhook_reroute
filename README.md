# # Webhook Re-Route

The Webhook Re-Route application is a simple service that accepts incoming HTTP POST requests and forwards them to a specified target URL. This can be useful when you have a webhook with a long URL that exceeds the character limit allowed by some services or when you want to add a layer of indirection to webhook URLs for easier maintenance.

## Features

- Receive incoming HTTP POST requests at a specified endpoint
- Forward incoming requests to a target URL with the original payload
- Respond with a custom status code and message to the caller

## Getting Started

### Prerequisites

- Golang 1.20.3 or later
- Docker (optional)

### Installation

1. Clone the repository:
   ```
   git clone https://github.com/itunzae/webhook_reroute.git
   ```
2. Change to the project directory:
   ```
   cd webhook_reroute
   ```

### Configuration

1. Create a `.env` file in the project root directory with the following content:
   ```
   TARGET_URL=https://your-target-url.example.com/path
   ```
   Replace `https://your-target-url.example.com/path` with the target URL to which you want to forward incoming webhook requests.

### Running the Application

#### Option 1: Run locally

1. Build and run the application:
   ```
   go build
   ./webhook-rewrite
   ```
2. The application will start, and the webhook re-route endpoint will be accessible at `http://localhost:8088/webhook`.

#### Option 2: Run using Docker

1. Build the Docker image:
   ```
   docker build -t webhook-rewrite .
   ```
2. Run a container using the built image:
   ```
   docker run -d -p 8088:8088 --name webhook-rewrite-container webhook-rewrite
   ```
3. The webhook re-route endpoint will be accessible at `http://localhost:8088/webhook`.

### Usage

To use the webhook re-route service, configure the webhook sender to send HTTP POST requests to `http://your-webhook-rewrite-host:8088/webhook`. The service will forward the requests to the target URL specified in the `.env` file.

## License

This project is licensed under the MIT License. See the `LICENSE` file for more information.
