package locust

type response struct {
	ClientName string
	RequestID  string
	Ping       int
}

type resultSet struct {
	count   int
	results map[string]int
}
