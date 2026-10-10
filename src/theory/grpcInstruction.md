# gRPC — краткая шпаргалка на примере Auth Service

## 1. Что такое gRPC

**gRPC** — технология удалённого вызова процедур (**RPC, Remote Procedure Call**).

Основная идея: одно приложение может вызвать функцию другого приложения почти так, как будто это обычный локальный метод.

Например, Shop вызывает:

```go
resp, err := authClient.Login(ctx, request)
```

Но фактически происходит сетевой запрос:

```text
Shop
  │
  │ LoginRequest
  │
  │ gRPC / HTTP/2
  ▼
Auth Service
  │
  │ Login(...)
  ▼
PostgreSQL
```

В нашем проекте:

- **Shop Service** предоставляет HTTP REST API клиенту;
- **Auth Service** — отдельное приложение;
- Shop общается с Auth Service только через **gRPC**;
- Auth Service имеет собственную PostgreSQL;
- формат сообщений описывается через **Protocol Buffers**.

---

# 2. REST и gRPC

Внешний клиент работает с Shop через обычный HTTP:

```text
Client
   │
   │ HTTP + JSON
   ▼
Shop Service
```

Например:

```http
POST /api/v1/auth
Content-Type: Application/json
```

```json
{
  "email": "user@mail.com",
  "password": "password"
}
```

Но внутри системы Shop обращается к Auth Service через gRPC:

```text
Client
   │
   │ HTTP/JSON
   ▼
Shop
   │
   │ gRPC/Protobuf
   ▼
Auth Service
   │
   ▼
Auth PostgreSQL
```

То есть REST и gRPC не конкурируют друг с другом. Они могут использоваться на разных уровнях одной системы.

В нашем случае:

```text
              public API
Client ───── HTTP/JSON ─────► Shop
                                │
                                │ internal API
                                │ gRPC/Protobuf
                                ▼
                           Auth Service
                                │
                                ▼
                           PostgreSQL
```

---

# 3. Почему gRPC удобно между сервисами

gRPC особенно хорошо подходит для взаимодействия между backend-сервисами.

Преимущества:

- строгий контракт API;
- автоматически генерируемые клиент и сервер;
- Protocol Buffers вместо ручного JSON;
- HTTP/2;
- поддержка streaming;
- типизация запросов и ответов;
- стандартные коды ошибок;
- удобно использовать в микросервисной архитектуре.

Например, вместо ручного формирования HTTP-запроса:

```go
http.Post(...)
```

мы можем написать:

```go
resp, err := authClient.Login(ctx, &pb.LoginRequest{
    Email:    email,
    Password: password,
})
```

---

# 4. Protocol Buffers

Для описания gRPC API используется файл `.proto`.

В нашем проекте:

```text
proto/auth.proto
```

Упрощённо:

```proto
syntax = "proto3";

package auth;

service AuthService {
    rpc Registration(RegistrationRequest)
        returns (RegistrationResponse);

    rpc Login(LoginRequest)
        returns (LoginResponse);

    rpc ValidateToken(ValidateTokenRequest)
        returns (ValidateTokenResponse);

    rpc ChangePassword(ChangePasswordRequest)
        returns (ChangePasswordResponse);

    rpc RestorePassword(RestorePasswordRequest)
        returns (RestorePasswordResponse);
}
```

`service` описывает доступные удалённые методы.

По смыслу это похоже на Go interface:

```go
type AuthService interface {
    Registration(...)
    Login(...)
    ValidateToken(...)
    ChangePassword(...)
    RestorePassword(...)
}
```

---

# 5. Message

Входные и выходные данные описываются через `message`.

Например:

```proto
message LoginRequest {
    string email = 1;
    string password = 2;
}

message LoginResponse {
    string accessToken = 1;
}
```

То есть RPC:

```proto
rpc Login(LoginRequest) returns (LoginResponse);
```

можно мысленно представить как:

```go
func Login(LoginRequest) LoginResponse
```

С учётом Go-контекста и ошибок реальный сгенерированный интерфейс будет ближе к:

```go
Login(
    context.Context,
    *LoginRequest,
) (*LoginResponse, error)
```

---

# 6. Зачем нужны номера полей

Например:

```proto
message User {
    string email = 1;
    string name = 2;
    string surname = 3;
}
```

Числа:

```text
1
2
3
```

— это **field numbers / tags**.

Они используются в бинарном формате Protocol Buffers для идентификации полей.

Это НЕ:

- порядок отображения;
- индекс массива;
- значение поля.

После публикации API номера полей нельзя бездумно менять.

Например, было:

```proto
string email = 1;
string name = 2;
```

Не следует потом делать:

```proto
string name = 1;
string email = 2;
```

Старые клиенты могут неправильно интерпретировать сообщения.

Если поле удаляется, хорошая практика — зарезервировать его номер:

```proto
reserved 2;
```

---

# 7. Генерация Go-кода

Из `.proto` генерируется Go-код.

Обычно используются:

```bash
protoc \
  --go_out=. \
  --go-grpc_out=. \
  proto/auth.proto
```

Нужны плагины:

```text
protoc-gen-go
protoc-gen-go-grpc
```

После генерации появляются структуры примерно такого вида:

```go
type LoginRequest struct {
    Email    string
    Password string
}
```

и gRPC client:

```go
type AuthServiceClient interface {
    Login(
        ctx context.Context,
        in *LoginRequest,
        opts ...grpc.CallOption,
    ) (*LoginResponse, error)
}
```

А также интерфейс сервера.

Главная идея:

```text
auth.proto
    │
    │ protoc
    ▼
generated Go code
    │
    ├── Messages
    ├── Client
    └── Server interfaces
```

Не нужно вручную писать сетевой протокол.

---

# 8. gRPC Server

Auth Service является gRPC-сервером.

Сначала создаём TCP listener:

```go
listener, err := net.Listen("tcp", ":50051")
```

Затем создаём gRPC server:

```go
grpcServer := grpc.NewServer()
```

Регистрируем реализацию:

```go
pbAuth.RegisterAuthServiceServer(
    grpcServer,
    authServer,
)
```

И запускаем:

```go
err = grpcServer.Serve(listener)
```

Полная схема:

```text
:50051
   │
   ▼
grpc.Server
   │
   ▼
AuthServer
   │
   ▼
UserService
   │
   ▼
UserRepository
   │
   ▼
PostgreSQL
```

---

# 9. Реализация gRPC Server

Сгенерированный код определяет контракт, а бизнес-логику пишем сами.

Например:

```go
func (auth *AuthServer) Login(
    ctx context.Context,
    req *pbAuth.LoginRequest,
) (*pbAuth.LoginResponse, error) {

    user, err := auth.userService.Authenticate(
        ctx,
        req.Email,
        req.Password,
    )
    if err != nil {
        return nil, toGRPCError(err)
    }

    token, err := auth.jwtProvider.GenerateAccessToken(user.ID)
    if err != nil {
        return nil, toGRPCError(err)
    }

    return &pbAuth.LoginResponse{
        AccessToken: token,
    }, nil
}
```

gRPC handler здесь выполняет роль транспортного слоя:

```text
gRPC request
     ↓
AuthServer
     ↓
Service
     ↓
Repository
```

Сам `AuthServer` не должен содержать SQL.

---

# 10. gRPC Client

Shop является gRPC-клиентом.

Сначала создаётся соединение:

```go
conn, err := grpc.NewClient(
    "localhost:50051",
    grpc.WithTransportCredentials(
        insecure.NewCredentials(),
    ),
)
```

После этого создаётся клиент:

```go
authClient := pbAuth.NewAuthServiceClient(conn)
```

Теперь можно вызывать удалённые методы:

```go
resp, err := authClient.Login(
    ctx,
    &pbAuth.LoginRequest{
        Email:    email,
        Password: password,
    },
)
```

С точки зрения Shop это выглядит как вызов метода.

Но внутри:

```text
Shop
 │
 │ protobuf serialization
 ▼
HTTP/2
 │
 ▼
TCP
 │
 ▼
Auth Service
 │
 │ protobuf deserialization
 ▼
Login(...)
```

---

# 11. Что происходит при вызове gRPC

Например:

```go
resp, err := authClient.Login(ctx, req)
```

Высокоуровнево:

```text
1. Go LoginRequest
        ↓
2. protobuf serialization
        ↓
3. gRPC framing
        ↓
4. HTTP/2
        ↓
5. network
        ↓
6. Auth gRPC Server
        ↓
7. protobuf deserialization
        ↓
8. AuthServer.Login()
        ↓
9. LoginResponse
        ↓
10. обратно тем же путём
```

Большую часть этого gRPC выполняет автоматически.

---

# 12. HTTP/2

gRPC работает поверх **HTTP/2**.

Упрощённый стек:

```text
Application
    │
    ▼
gRPC
    │
    ▼
Protocol Buffers
    │
    ▼
HTTP/2
    │
    ▼
TCP
```

HTTP/2 позволяет эффективно использовать одно соединение для нескольких запросов.

В обычном Go gRPC:

```go
grpc.NewServer()
```

и:

```go
pb.NewAuthServiceClient(...)
```

уже используют необходимую gRPC/HTTP2 инфраструктуру.

Не нужно самостоятельно реализовывать HTTP/2 handler.

---

# 13. JSON vs Protocol Buffers

REST часто использует:

```json
{
  "email": "user@mail.com",
  "password": "password"
}
```

gRPC обычно использует бинарный Protocol Buffers.

Мы в коде всё равно работаем с Go-структурой:

```go
&pbAuth.LoginRequest{
    Email:    "user@mail.com",
    Password: "password",
}
```

Но по сети она передаётся не как обычный JSON.

Схематично:

```text
Go struct
    ↓
protobuf.Marshal
    ↓
binary data
    ↓
network
    ↓
protobuf.Unmarshal
    ↓
Go struct
```

---

# 14. Context в gRPC

gRPC методы принимают:

```go
context.Context
```

Например:

```go
func (auth *AuthServer) Login(
    ctx context.Context,
    req *pbAuth.LoginRequest,
)
```

Shop передаёт:

```go
app.grpcAuthClient.Login(
    r.Context(),
    grpcReq,
)
```

Получается цепочка:

```text
HTTP request context
        ↓
Shop Handler
        ↓
gRPC Client
        ↓
AuthServer
        ↓
UserService
        ↓
Repository
        ↓
PostgreSQL
```

Это очень полезно.

Если клиент отменил HTTP-запрос, отмена может распространиться дальше по цепочке.

---

# 15. Deadline

Для межсервисных запросов желательно использовать timeout.

Например:

```go
ctx, cancel := context.WithTimeout(
    r.Context(),
    2*time.Second,
)
defer cancel()

resp, err := app.grpcAuthClient.ValidateToken(
    ctx,
    req,
)
```

Если Auth Service зависнет, Shop не будет ждать бесконечно.

Получим:

```text
Shop
 │
 ├── Auth ответил за < 2 sec → продолжаем
 │
 └── timeout → ошибка
```

Для production-сервисов deadlines практически обязательны.

---

# 16. Ошибки gRPC

В gRPC есть стандартные status codes.

Например:

```text
InvalidArgument
Unauthenticated
PermissionDenied
NotFound
AlreadyExists
Internal
Unavailable
DeadlineExceeded
```

В Go:

```go
status.Error(
    codes.Unauthenticated,
    "invalid email or password",
)
```

В нашем проекте repository и business logic не должны знать о gRPC.

Поэтому используется цепочка:

```text
Repository error
      ↓
Service error
      ↓
gRPC status
      ↓
Shop
      ↓
HTTP status
```

---

# 17. Ошибки Repository

Например PostgreSQL сообщает:

```text
unique violation
SQLSTATE 23505
```

Repository преобразует её в:

```go
repository.ErrConflict
```

Например:

```go
if pgErr.Code == "23505" {
    return repository.ErrConflict
}
```

Repository ничего не знает про:

```text
HTTP 409
gRPC AlreadyExists
```

Это правильно.

---

# 18. Service errors

Service работает с domain errors:

```go
var (
    ErrInvalidCredentials = errors.New(
        "invalid email or password",
    )

    ErrInvalidToken = errors.New(
        "invalid token",
    )

    ErrInvalidData = errors.New(
        "invalid data",
    )
)
```

Service также ничего не должен знать про HTTP.

---

# 19. Преобразование в gRPC errors

На границе gRPC:

```go
func toGRPCError(err error) error {
    switch {
    case errors.Is(err, ErrInvalidData):
        return status.Error(
            codes.InvalidArgument,
            "invalid input data",
        )

    case errors.Is(err, ErrInvalidCredentials):
        return status.Error(
            codes.Unauthenticated,
            "invalid email or password",
        )

    case errors.Is(err, repository.ErrNotFound):
        return status.Error(
            codes.NotFound,
            "user not found",
        )

    case errors.Is(err, repository.ErrConflict):
        return status.Error(
            codes.AlreadyExists,
            "user already exists",
        )

    default:
        return status.Error(
            codes.Internal,
            "internal server error",
        )
    }
}
```

Это транспортная граница:

```text
internal Application errors
            ↓
       gRPC errors
```

---

# 20. gRPC → HTTP

Shop получает gRPC error и преобразует его в HTTP.

Например:

```go
func grpcErrorToHTTP(err error) int {
    switch status.Code(err) {

    case codes.InvalidArgument:
        return http.StatusBadRequest

    case codes.Unauthenticated:
        return http.StatusUnauthorized

    case codes.PermissionDenied:
        return http.StatusForbidden

    case codes.NotFound:
        return http.StatusNotFound

    case codes.AlreadyExists:
        return http.StatusConflict

    default:
        return http.StatusInternalServerError
    }
}
```

Получается:

```text
Auth                        Shop

InvalidArgument     ───►    400
Unauthenticated     ───►    401
PermissionDenied    ───►    403
NotFound            ───►    404
AlreadyExists       ───►    409
Internal            ───►    500
```

---

# 21. Полный путь ошибки регистрации

Например пользователь регистрирует существующий email:

```text
POST /api/v1/register
        ↓
Shop Register
        ↓
gRPC Registration
        ↓
AuthServer
        ↓
UserService.Create
        ↓
UserRepository.Create
        ↓
PostgreSQL
        ↓
23505 unique violation
        ↓
repository.ErrConflict
        ↓
toGRPCError()
        ↓
codes.AlreadyExists
        ↓
gRPC
        ↓
Shop
        ↓
grpcErrorToHTTP()
        ↓
409 Conflict
```

Это хороший пример разделения ответственности между слоями.

---

# 22. Authentication и Authorization

Это разные понятия.

## Authentication

Вопрос:

> Кто ты?

В нашем проекте:

```text
email + password
        ↓
Auth Service
        ↓
bcrypt verification
        ↓
JWT
```

Метод:

```proto
rpc Login(LoginRequest)
    returns (LoginResponse);
```

---

## Authorization

Вопрос:

> Имеешь ли ты право выполнить этот запрос?

В текущем проекте авторизация простая:

```text
есть валидный JWT
        ↓
можно использовать protected API
```

Пока нет ролей:

```text
admin
user
manager
```

Поэтому фактически проверяется факт успешной аутентификации.

---

# 23. JWT + gRPC в нашем проекте

После регистрации или Login Auth возвращает JWT:

```text
Registration/Login
       ↓
Auth Service
       ↓
JWTProvider
       ↓
signed JWT
       ↓
Shop
       ↓
Client
```

Клиент затем отправляет:

```http
Authorization: Bearer <JWT>
```

Например:

```http
GET /api/v1/clients
Authorization: Bearer eyJhbGciOi...
```

---

# 24. Middleware

Shop не проверяет JWT самостоятельно.

Middleware достаёт header:

```go
authHeader := r.Header.Get("Authorization")
```

Проверяет формат:

```text
Bearer <token>
```

и отправляет token в Auth Service:

```go
resp, err := app.grpcAuthClient.ValidateToken(
    r.Context(),
    &pbAuth.ValidateTokenRequest{
        Token: token,
    },
)
```

Полная цепочка:

```text
Client
  │
  │ Authorization: Bearer JWT
  ▼
Shop
  │
  ▼
authJwtMiddleware
  │
  │ gRPC ValidateToken
  ▼
Auth Service
  │
  ▼
JWTProvider
  │
  ├── invalid ─────► gRPC Unauthenticated
  │                        ↓
  │                    HTTP 401
  │
  └── valid
       ↓
     userID
       ↓
Shop handler
```

Это ключевая часть архитектуры проекта.

---

# 25. Почему Shop сам не проверяет JWT

Технически Shop мог бы иметь secret key и самостоятельно делать:

```go
jwt.Parse(...)
```

Но по заданию верификация должна происходить через Auth Service.

Поэтому:

```text
НЕ:

Shop → jwt.Parse()


А:

Shop → gRPC → Auth → jwt.Parse()
```

Плюс такого подхода — логика аутентификации централизована.

Если сервисов станет несколько:

```text
Shop ───────┐
            │
Orders ─────┼──► Auth Service
            │
Delivery ───┘
```

они могут использовать единый сервис аутентификации.

---

# 26. Registration

HTTP:

```http
POST /api/v1/register
```

Shop создаёт:

```go
pbAuth.RegistrationRequest{
    Email:    req.Email,
    Name:     req.Name,
    Surname:  req.Surname,
    Phone:    req.Phone,
    Password: req.Password,
}
```

и вызывает:

```go
app.grpcAuthClient.Registration(...)
```

Auth Service:

```text
Registration
     ↓
validation
     ↓
bcrypt
     ↓
UserRepository.Create
     ↓
PostgreSQL
     ↓
JWTProvider
     ↓
AccessToken
```

---

# 27. Password hashing

Пароли нельзя хранить:

```text
password123
qwerty
123456
```

В Auth Service используется bcrypt:

```go
bcrypt.GenerateFromPassword(
    []byte(password),
    bcrypt.DefaultCost,
)
```

В PostgreSQL хранится hash:

```text
$2a$10$...
```

bcrypt включает соль в результат.

При Login:

```go
bcrypt.CompareHashAndPassword(
    []byte(user.PasswordHash),
    []byte(password),
)
```

То есть исходный пароль из hash обратно не извлекается.

---

# 28. Login

Цепочка:

```text
POST /api/v1/auth
        ↓
Shop
        ↓
gRPC Login
        ↓
Auth Service
        ↓
GetByEmail
        ↓
bcrypt.CompareHashAndPassword
        ↓
JWTProvider
        ↓
JWT
        ↓
Shop
        ↓
HTTP response
```

Ответ:

```json
{
  "access_token": "eyJhbGciOi..."
}
```

---

# 29. ChangePassword

Endpoint защищён middleware:

```text
PATCH /api/v1/auth/password
```

Сначала:

```text
JWT
 ↓
middleware
 ↓
gRPC ValidateToken
 ↓
OK
```

Затем handler вызывает:

```text
gRPC ChangePassword
```

Auth Service снова получает token и определяет пользователя.

После этого:

```text
new password
     ↓
bcrypt
     ↓
new password_hash
     ↓
UPDATE users
```

Повторная проверка JWT внутри `ChangePassword` полезна тем, что сам gRPC метод остаётся защищённым даже при прямом обращении к Auth Service.

---

# 30. RestorePassword

Endpoint:

```text
POST /api/v1/reset
```

Пользователь передаёт:

```json
{
  "email": "user@mail.com"
}
```

Shop вызывает:

```text
gRPC RestorePassword
```

Auth Service:

```text
email
  ↓
GetByEmail
  ↓
generatePassword()
  ↓
new plaintext password
  ↓
bcrypt
  ↓
UPDATE password_hash
  ↓
print plaintext password to console
```

Вывод в консоль:

```text
Password reset for user@mail.com.
New password: x9Km3QaP71Zd
```

Это заглушка реальной отправки email.

---

# 31. Отдельная БД Auth Service

Auth Service имеет собственное хранилище:

```text
Shop Service
    │
    ▼
shop-db
localhost:5432


Auth Service
    │
    ▼
users-db
localhost:5431
```

Это важно для разделения сервисов.

Auth Service владеет своими данными:

```text
users
password_hash
authentication data
```

Shop не должен напрямую читать таблицу пользователей Auth Service.

Правильный путь:

```text
Shop
  │
  │ gRPC
  ▼
Auth
  │
  ▼
users-db
```

А не:

```text
Shop ─────────────► users-db
```

---

# 32. Dependency Injection

Auth Service собирается примерно так:

```go
userRepo := postgres.NewPostgresUserRepository(db.Pool)

userService := service.NewUserService(userRepo)

jwtProvider := service.NewJWTProvider(secretKey)

authServer := service.NewAuthServer(
    userService,
    jwtProvider,
)
```

Получается:

```text
PostgresUserRepository
          │
          │ implements
          ▼
   UserRepository
          │
          │ injected
          ▼
     UserService
          │
          ▼
      AuthServer
```

Service зависит не от конкретного PostgreSQL repository, а от interface:

```go
type UserService struct {
    repo repository.UserRepository
}
```

Это позволяет заменить:

```text
PostgresUserRepository
```

например на:

```text
MockUserRepository
InMemoryUserRepository
```

без изменения `UserService`.

---

# 33. Dependency Injection gRPC Client

Shop также не создаёт Auth Client внутри каждого handler.

Плохо:

```go
func Login(...) {
    conn, _ := grpc.NewClient(...)
    client := pb.NewAuthServiceClient(conn)
}
```

Лучше один раз собрать приложение:

```go
type Application struct {
    grpcAuthClient pbAuth.AuthServiceClient
}
```

и затем использовать:

```go
app.grpcAuthClient.Login(...)
```

То есть соединение и зависимости создаются в composition root.

---

# 34. Unary RPC

Все методы нашего Auth Service сейчас являются **Unary RPC**.

Например:

```proto
rpc Login(LoginRequest)
    returns (LoginResponse);
```

Это:

```text
1 request
   ↓
server
   ↓
1 response
```

Аналог обычного request-response.

---

# 35. Streaming в gRPC

Кроме Unary RPC, gRPC поддерживает streaming.

## Server streaming

Один request → много responses:

```text
Client
  │ request
  ▼
Server
  │
  ├── response 1
  ├── response 2
  ├── response 3
  └── response 4
```

Пример:

```proto
rpc GetEvents(EventRequest)
    returns (stream Event);
```

---

## Client streaming

Много requests → один response:

```text
Client
  ├── request 1
  ├── request 2
  └── request 3
          ↓
        Server
          ↓
       response
```

---

## Bidirectional streaming

Обе стороны могут отправлять поток сообщений:

```text
Client ◄────────► Server
Client ◄────────► Server
Client ◄────────► Server
```

Например:

- chat;
- realtime telemetry;
- live updates.

Для нашего Auth Service streaming не нужен — unary RPC подходит лучше.

---

# 36. Metadata

У gRPC есть metadata — аналог HTTP headers.

Например можно передавать:

```text
authorization
request-id
trace-id
```

Пример:

```go
md := metadata.Pairs(
    "authorization",
    "Bearer "+token,
)

ctx := metadata.NewOutgoingContext(
    context.Background(),
    md,
)
```

После этого metadata можно читать на сервере.

В нашем проекте JWT сейчас передаётся непосредственно внутри protobuf message:

```proto
message ValidateTokenRequest {
    string token = 1;
}
```

Это допустимо для учебного проекта.

В более общем API authentication token часто передают через gRPC metadata.

---

# 37. Interceptors

У gRPC есть аналог HTTP middleware — **interceptors**.

В HTTP:

```text
Request
   ↓
Middleware
   ↓
Handler
```

В gRPC:

```text
RPC
 ↓
Interceptor
 ↓
gRPC Handler
```

Например можно сделать:

```go
grpc.NewServer(
    grpc.UnaryInterceptor(authInterceptor),
)
```

Interceptor может:

- проверять JWT;
- логировать запросы;
- измерять время;
- добавлять tracing;
- проверять permissions.

То есть:

```text
HTTP middleware ≈ gRPC interceptor
```

Это полезное соответствие для запоминания.

---

# 38. TLS

В нашем локальном проекте можно использовать:

```go
insecure.NewCredentials()
```

потому что взаимодействие происходит локально.

Для production:

```text
Shop
 │
 │ TLS
 ▼
Auth
```

желательно использовать TLS credentials.

`insecure.NewCredentials()` означает отсутствие TLS-защиты транспортного канала.

Это нормально для локальной разработки, но не является хорошим production default.

---

# 39. gRPC Health Check

Для production gRPC-сервиса полезно реализовать стандартный health service.

Тогда инфраструктура может спросить:

```text
Auth Service жив?
```

и получить:

```text
SERVING
```

Это используется:

- Docker;
- Kubernetes;
- load balancers;
- monitoring.

---

# 40. Graceful Shutdown

Вместо мгновенного завершения gRPC server:

```go
grpcServer.Stop()
```

часто лучше:

```go
grpcServer.GracefulStop()
```

Он позволяет текущим RPC завершиться перед остановкой сервиса.

Схема:

```text
SIGTERM
   ↓
stop accepting new RPC
   ↓
finish current RPC
   ↓
shutdown
```

---

# 41. Что стоит добавить в production

Для учебного проекта текущей реализации достаточно.

Для реального сервиса я бы добавил:

```text
Auth Service
├── config from environment
├── TLS
├── deadlines
├── logging interceptor
├── recovery interceptor
├── metrics
├── tracing
├── health check
├── graceful shutdown
├── refresh tokens
├── key rotation
└── rate limiting
```

В частности, вместо:

```go
secretKey := "secret"
```

использовать переменную окружения:

```text
JWT_SECRET
```

А вместо hardcoded:

```text
postgres://users:users@localhost:5431/users
```

использовать:

```text
AUTH_DB_URL
```

---

# 42. Главное архитектурное правило

gRPC — это **transport layer**, а не место для бизнес-логики.

Плохо:

```text
AuthServer.Login
    ├── SQL
    ├── bcrypt
    ├── JWT
    ├── validation
    └── everything else
```

Лучше:

```text
gRPC
 │
 ▼
AuthServer
 │
 ▼
Service
 │
 ├── JWTProvider
 │
 ▼
Repository
 │
 ▼
PostgreSQL
```

То есть каждый слой отвечает за своё.

---

# 43. Архитектура нашего проекта целиком

```text
                         CLIENT
                           │
                           │ HTTP/JSON
                           │
                           ▼
                  ┌─────────────────┐
                  │   SHOP SERVICE  │
                  │    :4000        │
                  └────────┬────────┘
                           │
              ┌────────────┴────────────┐
              │                         │
              ▼                         ▼
      authJwtMiddleware          Shop Service Layer
              │                         │
              │                         ▼
              │                  Shop Repository
              │                         │
              │                         ▼
              │                     shop-db
              │                     :5432
              │
              │ gRPC / HTTP2
              ▼
       ┌───────────────────┐
       │   AUTH SERVICE    │
       │      :50051       │
       └─────────┬─────────┘
                 │
                 ▼
            AuthServer
                 │
                 ▼
            UserService
            /         \
           ▼           ▼
    JWTProvider    UserRepository
                       │
                       ▼
                    users-db
                     :5431
```

---

# 44. Полная авторизация запроса

Например:

```http
GET /api/v1/clients
Authorization: Bearer eyJhbGciOi...
```

проходит:

```text
Client
  │
  │ GET /api/v1/clients
  │ Authorization: Bearer JWT
  ▼
Shop :4000
  │
  ▼
authJwtMiddleware
  │
  │ extract JWT
  ▼
AuthServiceClient.ValidateToken()
  │
  │ gRPC / HTTP2
  ▼
Auth Service :50051
  │
  ▼
JWTProvider.ValidateTokenAndGetUserID()
  │
  ├──────── invalid ─────────┐
  │                          │
  ▼                          ▼
userID               Unauthenticated
  │                          │
  │                          ▼
  │                       HTTP 401
  │
  ▼
ValidateTokenResponse
  │
  ▼
Shop middleware
  │
  ▼
Client Handler
  │
  ▼
Client Service
  │
  ▼
Client Repository
  │
  ▼
shop-db
  │
  ▼
HTTP 200
```

---

# 45. Полная регистрация

```text
POST /api/v1/register
         │
         ▼
   Shop Register
         │
         │ RegistrationRequest
         ▼
       gRPC
         │
         ▼
 AuthServer.Registration
         │
         ▼
 UserService.Create
         │
         ├── validation
         │
         ├── UUID
         │
         └── bcrypt
                 │
                 ▼
          UserRepository
                 │
                 ▼
             users-db
                 │
                 ▼
           JWTProvider
                 │
                 ▼
              JWT
                 │
                 ▼
             gRPC
                 │
                 ▼
              Shop
                 │
                 ▼
        HTTP 201 Created

{
    "access_token": "..."
}
```

---

# 46. Короткая шпаргалка

```text
.proto
    ↓
описывает контракт API

protoc
    ↓
генерирует Go client/server code

grpc.Server
    ↓
принимает RPC

AuthServiceClient
    ↓
вызывает RPC

message
    ↓
структура request/response

protobuf
    ↓
бинарная сериализация

HTTP/2
    ↓
транспорт gRPC

context
    ↓
cancel / deadline / metadata

status.Error
    ↓
gRPC error

codes.Unauthenticated
    ↓
стандартный gRPC status code

interceptor
    ↓
аналог middleware

metadata
    ↓
аналог headers
```

---

# 47. Самое главное для нашего проекта

Нужно хорошо понимать следующую цепочку:

```text
                 HTTP
Client ─────────────────────► Shop
                               │
                               │ gRPC
                               ▼
                         Auth Service
                               │
                               │ SQL
                               ▼
                          PostgreSQL
```

Shop отвечает за внешний REST API.

Auth Service отвечает за:

```text
Registration
Authentication
JWT generation
JWT validation
Password change
Password restore
```

Shop **не обращается напрямую** к Auth PostgreSQL.

Shop обращается:

```text
Shop
 ↓
AuthServiceClient
 ↓
gRPC / HTTP2
 ↓
AuthServer
 ↓
Service
 ↓
Repository
 ↓
PostgreSQL
```

А авторизация protected endpoints происходит:

```text
HTTP request
      ↓
middleware
      ↓
Authorization: Bearer JWT
      ↓
gRPC ValidateToken
      ↓
Auth Service
      ↓
valid?
 ┌────┴─────┐
 │          │
no         yes
 │          │
401      handler
```

Именно эта схема является центральной идеей gRPC в текущем проекте.