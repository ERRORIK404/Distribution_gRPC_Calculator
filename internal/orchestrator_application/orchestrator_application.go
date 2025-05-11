package orchestrator_application

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	db "github.com/ERRORIK404/Distribution_gRPC_Calculator/database"
	agent "github.com/ERRORIK404/Distribution_gRPC_Calculator/internal/agent_application"
	conf "github.com/ERRORIK404/Distribution_gRPC_Calculator/pkg/config"
	conv "github.com/ERRORIK404/Distribution_gRPC_Calculator/pkg/converter_to_RPN"
	hashing "github.com/ERRORIK404/Distribution_gRPC_Calculator/pkg/hashing"
	locerr "github.com/ERRORIK404/Distribution_gRPC_Calculator/pkg/local_errors"
	pb "github.com/ERRORIK404/Distribution_gRPC_Calculator/pkg/proto"
	structs "github.com/ERRORIK404/Distribution_gRPC_Calculator/pkg/structs"
	tokenezation "github.com/ERRORIK404/Distribution_gRPC_Calculator/pkg/tokenezation"
	validator "github.com/ERRORIK404/Distribution_gRPC_Calculator/pkg/validator"
	"google.golang.org/grpc"
)

var (
	config = conf.LoadConfig()
	tasksMap = structs.SafeTaskMap{Task_map: make(map[int32]structs.Task)}
	expressionsMap = structs.SafeExpressionMap{Expression_map: make(map[int32]structs.Expression)}
)

type contextKey string
const loginKey contextKey = "login"

// Структура дерева нужна чтобы конвертировать RPN в форму удобную для подсчета выражения с использованием concurrency кода
type ExprNode struct {
	Left  *ExprNode
	Right *ExprNode
	Op    string
	Value float64
}

// Конвертируем RPN запись в дерево для того, чтобы можно было асинхронно подсчитывать ветви
func RpnToTree(tokens []string) (*ExprNode, error) {
	stack := []*ExprNode{}

	for _, token := range tokens {
		if num, err := strconv.ParseFloat(token, 64); err == nil {
			// Если токен — число, создаем лист и добавляем в стек
			stack = append(stack, &ExprNode{Value: num})
		} else {
			// Если токен — оператор, извлекаем два операнда из стека
			if len(stack) < 2 {
				return nil, fmt.Errorf("недостаточно операндов для оператора %s", token)
			}

			// Извлекаем правый и левый операнды
			right := stack[len(stack)-1]
			left := stack[len(stack)-2]
			stack = stack[:len(stack)-2]

			// Создаем новый узел и добавляем его в стек
			node := &ExprNode{Op: token, Left: left, Right: right}
			stack = append(stack, node)
		}
	}

	if len(stack) != 1 {
		return nil, fmt.Errorf("некорректное выражение")
	}

	return stack[0], nil
}

// Функция для асинхронного подсчета ветвей
func (n *ExprNode) EvaluateParallel() float64 {
	if n.Left == nil && n.Right == nil {
		// Еслт это лист, возвращаем значение
		return n.Value
	}

	// Каналы для получения результатов
	leftChan := make(chan float64, 15)
	rightChan := make(chan float64, 15)

	// Запускаем вычисление левого поддерева
	go func() {
		leftChan <- n.Left.EvaluateParallel()
	}()

	// Запускаем вычисление правого поддерева
	go func() {
		rightChan <- n.Right.EvaluateParallel()
	}()

	// Ждем результаты
	left := <-leftChan
	right := <-rightChan

	// Создаем задачу, и ждем исполнения
	log.Println("CREATE TASK")
	switch n.Op {
		case "+":
			task := structs.NewTask(left, right, "+", config.TIME_ADDITION_MS)
			tasksMap.Write(task)
			res := <- tasksMap.Read(task.Id).Chan
			tasksMap.Delete(task.Id)
			return res
		case "-":
			task := structs.NewTask(left, right, "-", config.TIME_SUBTRACTION_MS)
			tasksMap.Write(task)
			res := <- tasksMap.Read(task.Id).Chan
			tasksMap.Delete(task.Id)
			return res
		case "*":
			task := structs.NewTask(left, right, "*", config.TIME_MULTIPLICATIONS_MS)
			tasksMap.Write(task)
			res := <- tasksMap.Read(task.Id).Chan
			tasksMap.Delete(task.Id)
			return res
		case "/":
			if right == 0 {
				return left
			}
			task := structs.NewTask(left, right, "/", config.TIME_DIVISIONS_MS)
			tasksMap.Write(task)
			res := <- tasksMap.Read(task.Id).Chan
			tasksMap.Delete(task.Id)
			return res
		default:
			return 0
	}
}

func Validator(expression string) (*ExprNode, error) {
	    // Удаляем пробелы из выражения
    expression = strings.ReplaceAll(expression, " ", "")

    // Проверка на пустоту
    if expression == "" {
        return nil, locerr.ErrEmptyExpression
    }

    // Проверка на правильность скобок
    if !conv.IsValidParentheses(expression) {
        return nil,locerr.ErrIncorrectBracketPlacement
    }

    // Преобразование выражения в RPN
    rpn, err := conv.InfixToRPN(expression)
    if err != nil {
        return nil, err
    }

	result, err := RpnToTree(rpn)
    if err != nil {
        return nil, err
    }
	return result, nil
}

// Хендлер для получчения нового выражения
func (h *AuthHandler) Accept_expression_handler(w http.ResponseWriter, r *http.Request) {
	login := r.Context().Value(loginKey).(string)

	expression := structs.NewExpression()
	type message struct{Message string `json:"expression"`}
	var str message

    defer r.Body.Close()
    err := json.NewDecoder(r.Body).Decode(&str)
    if err != nil {
        http.Error(w, err.Error(), http.StatusUnprocessableEntity)
        return
    }

	expression.Expression = str.Message
	expressionsMap.Write(expression)
	result, err := Validator(expression.Expression)
	if err != nil {
		expression.Status = "error"
		if err == locerr.ErrIncorrectBracketPlacement || err == locerr.ErrEmptyExpression{
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	type id struct{Id int32 `json:"id"`}
	json.NewEncoder(w).Encode(id{Id:expression.Id})
	h.db.AddHistoryEntry(expression.Id, login, 0, expression.Expression, "pending")
	// Запускаем вычисление выражения в отдельном потоке
    go func(){
		res := result.EvaluateParallel()
		expression.Result = fmt.Sprintf("%f", res)
		expression.Status = "success"
		h.db.UpdateHistoryEntry(expression.Id, login, expression.Expression, res, expression.Status)
		expressionsMap.Delete(expression.Id)
    }()
}

type Server struct {
	pb.UnimplementedOrchestratorServiceServer
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) GetTask(ctx context.Context, in *pb.EmptyMessage) (*pb.ServerResponse, error) {
	task, err := tasksMap.Get_does_not_have_result()
	if err != nil{
		return nil, status.Errorf(codes.NotFound, "operation failed: %v", err)
	}
	return &pb.ServerResponse{Task: &pb.Task{Id: int32(task.Id), Operand1: float32(task.Operand1), Operand2: float32(task.Operand2), Operation: task.Operation, OperationTime: int32(task.Operation_time), Status: task.Status}}, nil
}

func (s *Server) SendResult(ctx context.Context, in *pb.ResultRequest) (*pb.EmptyMessage, error) {
	res := in.Result
	tasksMap.Write_result(structs.Result{Id: res.Id, Result: float64(res.Result)})
	return &pb.EmptyMessage{}, nil
}

type AuthHandler struct {
    db *db.DB
}

func NewAuthHandler(db *db.DB) *AuthHandler {
    return &AuthHandler{db: db}
}

func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        cookie, err := r.Cookie("token")
        if err != nil {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        log.Println("Cookie found")

        tokenString := cookie.Value
        login, err := tokenezation.CheckToken(tokenString, config.JWT_SECRET);
        if  err != nil  {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        ctx := context.WithValue(r.Context(), loginKey, login)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

type UsersData struct {
	Login string
	Password string
}

func (h *AuthHandler) Register_user(w http.ResponseWriter, r *http.Request){
	var user UsersData
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err = validator.ValidateLogin(user.Login)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	hashpassword, err := hashing.HashPassword(user.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Println("Create user login: ", user.Login)
	err = h.db.CreateUser(user.Login, hashpassword)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *AuthHandler) Login_user(w http.ResponseWriter, r *http.Request) {
    var user UsersData
    err := json.NewDecoder(r.Body).Decode(&user)
    if err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    // Получаем пользователя из БД
    userFromDB, err := h.db.GetUser(user.Login)
    if err != nil {
        http.Error(w, "User not found", http.StatusUnauthorized)
        return
    }

    // Сравниваем исходный пароль с хешем из БД
    if !hashing.CheckPassword(userFromDB.PasswordHash, user.Password) {
        http.Error(w, "Invalid password", http.StatusUnauthorized)
        return
    }
	log.Println("Login user login: ", user.Login)
    // Генерируем токен
    tokenString, err := tokenezation.GenerateToken(user.Login, config.JWT_SECRET)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Устанавливаем куки
    http.SetCookie(w, &http.Cookie{
        Name:     "token",
        Value:    tokenString,
        Path:     "/",
        HttpOnly: true,
        Secure:   true,
        Expires:  time.Now().Add(10 * time.Minute),
    })

    w.WriteHeader(http.StatusOK)
}
// Хендлер для получения истории всех выражений
func (h *AuthHandler) Get_all_expressions_handler(w http.ResponseWriter, r *http.Request) {
	login := r.Context().Value(loginKey).(string)


    expressions := make([]structs.Expression, 0)
	history, err := h.db.GetUserHistoryByLogin(login)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for _, expression := range history {
		expressions = append(expressions, structs.Expression{Id: int32(expression.ID), Expression: expression.Expression, Result: fmt.Sprintf("%f", expression.Result), Status: expression.Status})
	}
    

    jsonData, err := json.Marshal(expressions)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.Write(jsonData)
	log.Println("Get all data from map")
}

//Хендре для получения выражения по его индентификатору
func (h *AuthHandler) Get_expression_by_id_handler(w http.ResponseWriter, r *http.Request) {
	login := r.Context().Value(loginKey).(string)
    var id int32
	n, err := fmt.Sscanf(r.URL.Path, "/api/v1/expressions/%d", &id)
	if n != 1 || err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	expression, err := h.db.GetHistoryEntryByID(id, login)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	expressionForResponse := structs.Expression{Id: int32(expression.ID), Expression: expression.Expression, Result: fmt.Sprintf("%f", expression.Result), Status: expression.Status}
    jsonData, err := json.Marshal(expressionForResponse)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.Write(jsonData)
}


func RunServer(){
	lis, err := net.Listen("tcp", "localhost:8081")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	orchestratorServer := NewServer()
	pb.RegisterOrchestratorServiceServer(grpcServer, orchestratorServer)
	go func(){
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()
	log.Println("Start grpc server")

	dbInstance, err := db.InitDB()
	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
	log.Println("Start db")

	authHandler := NewAuthHandler(dbInstance)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/auth/register", authHandler.Register_user)
	mux.HandleFunc("/api/v1/auth/login", authHandler.Login_user)
	mux.Handle("/api/v1/expressions/", AuthMiddleware(http.HandlerFunc(authHandler.Get_expression_by_id_handler)))
	mux.Handle("/api/v1/expressions", AuthMiddleware(http.HandlerFunc(authHandler.Get_all_expressions_handler)))
	mux.Handle("/api/v1/calculate", AuthMiddleware(http.HandlerFunc(authHandler.Accept_expression_handler)))
    // mux.HandleFunc("/internal/task", Assigned_task_handler)
	
	go func(){http.ListenAndServe(":8080", mux)}()
	log.Println("Start http server")
	for i := 0; i < config.COMPUTING_POWER; i++ {
		go func() {
			if agent.AgentRun() != nil{
				fmt.Println("Agent kill")
			}
		}()
	}
	log.Println("Start agents")
	for {
		time.Sleep(time.Minute * 10)
	}
}
