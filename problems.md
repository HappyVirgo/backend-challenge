## Bugs Found

### 1. Race Condition Bug
The fetchData function writes to a shared map without proper synchronization in multiple goroutines. This can lead to race conditions and data corruption when multiple departments are being processed concurrently. The solution is to use a sync.Mutex or sync.Map to protect access to the shared result map, or alternatively collect results in individual goroutines and combine them safely in the main thread.

### 2. Error Handling in POST Endpoint
The POST endpoint returns HTTP 500 for all errors, including client validation failures (missing email, department, or invalid results). This makes it impossible for clients to properly handle different types of errors and retries. The solution is to return appropriate HTTP status codes (400 for validation errors, 500 for server errors) and structured error responses.

## Future Development Challenges

### Hard to Maintain Survey Structure
The survey structure (exactly 4 results) is hardcoded throughout the codebase with magic numbers. This makes it difficult to modify the survey structure or add new questions in the future. Store the survey structure in a configuration that can be easily modified and validate against that structure.

## Code Maintainability Improvements

### AWS Session Management
A new AWS session and DynamoDB client is created for each request and goroutine. This is inefficient and can lead to connection pool exhaustion under load. Initialize these clients once at the Lambda container level and reuse them across requests.

## Scalability Concerns

### Inefficient DynamoDB Operations
Using DynamoDB Scan operations to fetch department results will become increasingly slow and expensive as the dataset grows. This is inefficient and will hit DynamoDB throughput limits at scale. Implement a Global Secondary Index on the department field and use Query operations instead of Scan, or consider implementing data aggregation during write operations.