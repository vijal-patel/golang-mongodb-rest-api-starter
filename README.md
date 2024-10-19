# golang-mongodb-rest-api-starter

This a work in progress REST API starter kit with the following features:

- **Swagger integration** for API documentation
- **RBAC** using Casbin for access control
- **Rate limiting**
- **Logging** with request ID, stack trace on error, and log level manipulation
- **User validation** via OTP and hCaptcha
- **Graceful shutdown** for clean resource management
- **API versioning**
- **MongoDB support**
- **Metrics** for monitoring
- **Hcaptcha** to prevent bots
- **Trace ID injection**

### 1. **Swagger Integration**

Swagger is enabled to provide interactive API documentation. It helps developers understand available endpoints and their request/response structures.

### 2. **Logging**

A custom logging wrapper adds:

- **Request ID** for tracking across multiple services.
- **Formatted logging** for improved readability.
- **Error stack trace** on failure to aid debugging.
- API to **change log levels** dynamically for fine-tuning logs at runtime.

### 3. **Graceful Shutdown**

The application supports graceful shutdown, ensuring that all in-flight requests are completed before shutting down services.

### 4. **RBAC with Casbin**

Role-Based Access Control (RBAC) is implemented using **Casbin**, enabling access control based on user roles and permissions.

### 5. **Authentication and User Validation**

- **OTP (One-Time Password) Endpoint**: Generates and validates OTPs for user authentication.
- **Resend OTP Endpoint**: Allows resending OTP for users.
- **hCaptcha Integration**: Protects certain endpoints from automated abuse via hCaptcha.

### 6. **Rate Limiting**

- **Login** and **register** endpoints are rate-limited by either email address or IP to prevent brute force attacks.

### 7. **API Versioning**

The application supports multiple API versions for backward compatibility.


## Environment Variables

View the .env.dev.sample file for a sample

## Running the Application

1. **Install Dependencies**
   ```bash
   go mod tidy
   ```
1. **Run**
   ```bash
   make run
   ```

## Build the application

```bash
  ko build cmd/main.go
```

Alternatively use docker or create a binary

## Inspired by
- https://github.com/google/exposure-notifications-server
- https://github.com/mattermost/mattermost/tree/master/api
- https://github.com/ardanlabs/service

