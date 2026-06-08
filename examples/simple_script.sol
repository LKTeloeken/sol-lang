rise Fibonacci {
    private int actual;
    private int next;

    glow() {
        this.actual = 0;
        this.next = 1;
    }

    public ray getActual() int {
        emit this.actual;
    }

    public ray getNext() int {
        emit this.next;
    }

    public ray next() void {
        var temp int = this.actual;
        this.actual = this.next;
        this.next = temp + this.next;
    }
}

var fibonacci Fibonacci = new Fibonacci();

var nums [int] = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10];

for each n in nums {

    Console.print(fibonacci.getNext());
    fibonacci.next();
}
