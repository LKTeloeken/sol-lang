rise HandleFile {
    private string filePath;

    glow(string filePath) {
        this.filePath = filePath;
    }

    public ray getFileContent() string {
        var content string = File.read(this.filePath);
        emit content;
    }

    public ray writeFileContent(string content) void {
        File.write(this.filePath, content);
    }
}

var handleFile HandleFile = new HandleFile("example.txt");

handleFile.writeFileContent("Hello, World!");
Console.print(handleFile.getFileContent());