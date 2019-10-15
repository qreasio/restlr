package middleware

// ApiKey is a function that will check and validate api key from every http request header
// Only request that has prefix as specified in allowedPaths will be skipped
// If the api key is not valid, will return 401 status
