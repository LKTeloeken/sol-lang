File.write("output.txt", "Hello from SOL\n");
File.append("output.txt", "more\n");
if (File.exists("output.txt")) {
    var conteudo string = File.read("output.txt");
    Console.print(conteudo);
}
