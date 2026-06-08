var i int = 0;
while (i < 3) {
    i = i + 1;
}

rise Demo {
    glow() {}

    public ray fail() {
        flare "test error";
    }
}

var d Demo = new Demo();

try {
    d.fail();
} catch (erro) {
    Console.print("caught: " + erro);
}
