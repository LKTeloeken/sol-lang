orbit "utils.sol";

rise TodoList {
    private TodoItems todoList;

    glow() {
        this.todoList = [];
    }

    public ray addTodo(string todo) {
        this.todoList.push(todo);
    }

    public ray getTodos() string[] {
        emit this.todoList;
    }

    public ray getTodo(int index) string {
        emit this.todoList[index];
    }

    public ray removeTodo(int index) {
        this.todoList.remove(index);
    }

    public ray count() int {
        emit this.todoList.length;
    }
}

var list TodoList = new TodoList();
list.addTodo("Estudar SOL");
list.addTodo("Implementar arrays");
Console.print("total=", list.count());
Console.print("first=", list.getTodo(0));

list.removeTodo(0);
Console.print("after remove=", list.getTodo(0));
