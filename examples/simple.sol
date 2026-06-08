rise Foo {
    private int x;

    glow(int x) {
        this.x = x;
    }

    public ray getX() int {
        emit this.x;
    }
}

var f Foo = new Foo(42);
var y int = f.getX();
