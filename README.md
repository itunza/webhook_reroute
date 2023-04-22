## Webhook Reroute

This is a simple Go application that receives incoming webhook POST requests and forwards them to another URL. The application supports both JSON and form data, including form URL-encoded data.

## Getting Started

### Prerequisites

- Go 1.20.3 or later
- Set up your environment variables in a `.env` file

### Environment Variables

Create a `.env` file in the project root with the following variables:


Replace `https://example.com/your-webhook-url` with the actual destination URL you want to forward the webhook requests to.

### Build and Run

To build the application, run the following command:


To start the server, run:


The server will start listening on port 8088.

## Endpoints

- `/webhook`: Accepts incoming webhook POST requests and forwards them to the destination URL specified in the `SUPPLIER_URL` environment variable. Supports both JSON and form data, including form URL-encoded data.

- `/add-url`: Accepts a URL via a form submission and returns a new URL that can be used to forward incoming webhook requests to the specified URL.

- `/`: Displays a simple form to add a new webhook URL.


## License

This project is licensed under the MIT License. See the `LICENSE` file for more information.

