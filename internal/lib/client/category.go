package client

const API_URL = "http://localhost:8090"

// func CategoryList(c *http.Client) category.ListResponse {
// 	req, err := c.Get(API_URL + "/categories")
// 	if err != nil {
// 		log.Fatalf("Error creating get request: %v", err)
// 	}
// 	defer req.Body.Close()

// 	resp, err := c.Do(req)
// 	if err != nil {
// 		log.Fatalf("Error sending get request: %v", err)
// 	}
// 	defer resp.Body.Close()

// 	var list category.ListResponse
// 	err = json.NewDecoder(resp.Body).Decode(&list)
// 	if err != nil {
// 		log.Fatalf("Error deconding list response: %v", err)
// 	}

// 	return list
// }
