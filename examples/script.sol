Console.print("Hello, SOL");

var nums int[] = [1, 2, 3];
for each n in nums {
    Console.print("n=" + n);
}

rise Greeter {
    private string name;

    glow(string name) {
        this.name = name;
    }

    public ray hello() {
        Console.print("Hi, " + this.name);
    }
}

var g Greeter = new Greeter("Ana");
g.hello();

rise Failer {
    glow() {}

    public ray boom() {
        flare "falhou";
    }
}

var f Failer = new Failer();

try {
    f.boom();
} catch (e) {
    Console.print("erro: " + e);
}
