# 多语言客户端代码示例

## Go客户端

```go
package apiclient

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
)

const BaseURL = "https://api.example.com/v1"

// Client API客户端
type Client struct {
    HTTPClient *http.Client
    Token      string
}

// NewClient 创建新客户端
func NewClient(token string) *Client {
    return &Client{
        HTTPClient: &http.Client{Timeout: 30 * time.Second},
        Token:      token,
    }
}

// User 用户模型
type User struct {
    ID        int       `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    Status    string    `json:"status"`
    CreatedAt time.Time `json:"createdAt"`
}

// GetUsers 获取用户列表
func (c *Client) GetUsers(offset, limit int) ([]User, int, error) {
    url := fmt.Sprintf("%s/users?offset=%d&limit=%d", BaseURL, offset, limit)

    req, _ := http.NewRequest("GET", url, nil)
    req.Header.Set("Authorization", "Bearer "+c.Token)

    resp, err := c.HTTPClient.Do(req)
    if err != nil {
        return nil, 0, err
    }
    defer resp.Body.Close()

    var result struct {
        Data   []User `json:"data"`
        Total  int    `json:"total"`
    }

    json.NewDecoder(resp.Body).Decode(&result)
    return result.Data, result.Total, nil
}

// CreateUser 创建用户
func (c *Client) CreateUser(name, email string) (*User, error) {
    data := map[string]string{"name": name, "email": email}
    body, _ := json.Marshal(data)

    req, _ := http.NewRequest("POST", BaseURL+"/users", bytes.NewBuffer(body))
    req.Header.Set("Authorization", "Bearer "+c.Token)
    req.Header.Set("Content-Type", "application/json")

    resp, err := c.HTTPClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var user User
    json.NewDecoder(resp.Body).Decode(&user)
    return &user, nil
}
```

## JavaScript/TypeScript客户端

```typescript
// api-client.ts
interface User {
  id: number;
  name: string;
  email: string;
  status: string;
  createdAt: string;
}

interface PaginatedResponse<T> {
  data: T[];
  total: number;
  offset: number;
  limit: number;
}

class ApiClient {
  private baseURL: string;
  private token: string;

  constructor(baseURL: string, token: string) {
    this.baseURL = baseURL;
    this.token = token;
  }

  private async request<T>(
    method: string,
    path: string,
    body?: any
  ): Promise<T> {
    const response = await fetch(`${this.baseURL}${path}`, {
      method,
      headers: {
        'Authorization': `Bearer ${this.token}`,
        'Content-Type': 'application/json',
      },
      body: body ? JSON.stringify(body) : undefined,
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error.message);
    }

    return response.json();
  }

  async getUsers(
    offset: number = 0,
    limit: number = 20
  ): Promise<PaginatedResponse<User>> {
    return this.request<PaginatedResponse<User>>(
      'GET',
      `/users?offset=${offset}&limit=${limit}`
    );
  }

  async createUser(name: string, email: string): Promise<User> {
    return this.request<User>('POST', '/users', { name, email });
  }

  async getUser(id: number): Promise<User> {
    return this.request<User>('GET', `/users/${id}`);
  }

  async updateUser(id: number, data: Partial<User>): Promise<User> {
    return this.request<User>('PATCH', `/users/${id}`, data);
  }

  async deleteUser(id: number): Promise<void> {
    await this.request<void>('DELETE', `/users/${id}`);
  }
}

// 使用示例
const client = new ApiClient('https://api.example.com/v1', 'your_token');

const users = await client.getUsers(0, 20);
console.log(`Total: ${users.total}`);

const newUser = await client.createUser('张三', 'zhangsan@example.com');
console.log(`Created: ${newUser.name}`);
```

## Python客户端

```python
import requests
from typing import List, Dict, Optional

class ApiClient:
    def __init__(self, base_url: str, token: str):
        self.base_url = base_url
        self.token = token
        self.session = requests.Session()
        self.session.headers.update({
            'Authorization': f'Bearer {token}',
            'Content-Type': 'application/json'
        })

    def _request(self, method: str, path: str, **kwargs):
        url = f"{self.base_url}{path}"
        response = self.session.request(method, url, **kwargs)

        if not response.ok:
            error = response.json()
            raise Exception(error['error']['message'])

        return response.json() if response.content else None

    def get_users(self, offset: int = 0, limit: int = 20) -> Dict:
        return self._request('GET', f'/users?offset={offset}&limit={limit}')

    def create_user(self, name: str, email: str) -> Dict:
        return self._request('POST', '/users', json={'name': name, 'email': email})

    def get_user(self, user_id: int) -> Dict:
        return self._request('GET', f'/users/{user_id}')

    def update_user(self, user_id: int, **kwargs) -> Dict:
        return self._request('PATCH', f'/users/{user_id}', json=kwargs)

    def delete_user(self, user_id: int) -> None:
        self._request('DELETE', f'/users/{user_id}')

# 使用示例
client = ApiClient('https://api.example.com/v1', 'your_token')

users = client.get_users(offset=0, limit=20)
print(f"Total: {users['total']}")

new_user = client.create_user('张三', 'zhangsan@example.com')
print(f"Created: {new_user['name']}")
```

## Java客户端

```java
import com.fasterxml.jackson.databind.ObjectMapper;
import java.net.http.*;
import java.net.URI;
import java.util.List;
import java.util.Map;

public class ApiClient {
    private final String baseURL;
    private final String token;
    private final HttpClient httpClient;
    private final ObjectMapper objectMapper;

    public ApiClient(String baseURL, String token) {
        this.baseURL = baseURL;
        this.token = token;
        this.httpClient = HttpClient.newHttpClient();
        this.objectMapper = new ObjectMapper();
    }

    public PaginatedResponse<User> getUsers(int offset, int limit) throws Exception {
        String url = String.format("%s/users?offset=%d&limit=%d", baseURL, offset, limit);

        HttpRequest request = HttpRequest.newBuilder()
            .uri(URI.create(url))
            .header("Authorization", "Bearer " + token)
            .GET()
            .build();

        HttpResponse<String> response = httpClient.send(request,
            HttpResponse.BodyHandlers.ofString());

        return objectMapper.readValue(response.body(),
            objectMapper.getTypeFactory().constructParametricType(
                PaginatedResponse.class, User.class));
    }

    public User createUser(String name, String email) throws Exception {
        Map<String, String> data = Map.of("name", name, "email", email);
        String json = objectMapper.writeValueAsString(data);

        HttpRequest request = HttpRequest.newBuilder()
            .uri(URI.create(baseURL + "/users"))
            .header("Authorization", "Bearer " + token)
            .header("Content-Type", "application/json")
            .POST(HttpRequest.BodyPublishers.ofString(json))
            .build();

        HttpResponse<String> response = httpClient.send(request,
            HttpResponse.BodyHandlers.ofString());

        return objectMapper.readValue(response.body(), User.class);
    }
}

class User {
    public int id;
    public String name;
    public String email;
    public String status;
    public String createdAt;
}

class PaginatedResponse<T> {
    public List<T> data;
    public int total;
    public int offset;
    public int limit;
}
```

## 使用示例

### Go
```go
client := apiclient.NewClient("your_token")
users, total, _ := client.GetUsers(0, 20)
fmt.Printf("Found %d users\n", total)
```

### JavaScript
```javascript
const client = new ApiClient('https://api.example.com/v1', 'your_token');
const users = await client.getUsers(0, 20);
console.log(`Found ${users.total} users`);
```

### Python
```python
client = ApiClient('https://api.example.com/v1', 'your_token')
users = client.get_users(offset=0, limit=20)
print(f"Found {users['total']} users")
```

### Java
```java
ApiClient client = new ApiClient("https://api.example.com/v1", "your_token");
PaginatedResponse<User> users = client.getUsers(0, 20);
System.out.println("Found " + users.total + " users");
```
