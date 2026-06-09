var parts string[] = String.split("one,two,three", ",");
for each p in parts {
    Console.print(String.trim(p));
}

var x float = Math.max(1.5, 2.5);
Console.print("max=", x);

var n int = Math.floor(3.9);
Console.print("floor=", n);
