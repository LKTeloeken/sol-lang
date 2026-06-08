rise User {
    private string name;
    private int age;

    glow(string username, int age) {
        this.name = username;
        this.age = age;
    }

    public ray getName() string {
        emit this.name;
    }

    public ray getAge() int {
        emit this.age;
    }
}

var user User = new User("John", 20);

var name string = user.getName();
var age int = user.getAge();

Console.print(name);
Console.print(age);

var test int = "test";