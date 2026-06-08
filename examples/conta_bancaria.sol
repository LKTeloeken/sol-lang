rise ContaBancaria {

    private float saldo;
    private string titular;

    glow(string titular, float saldoInicial) {
        this.titular = titular;
        this.saldo   = saldoInicial;
    }

    public ray depositar(float valor) {
        if (valor <= 0) {
            flare "Deposit value must be positive";
        }
        this.saldo = this.saldo + valor;
    }

    public ray sacar(float valor) {
        if (valor > this.saldo) {
            flare "Insufficient balance";
        }
        this.saldo = this.saldo - valor;
    }

    public ray getSaldo() {
        emit this.saldo;
    }

    public ray getTitular() {
        emit this.titular;
    }
}

var conta ContaBancaria = new ContaBancaria("Ana Silva", 1000.00);

conta.depositar(250.00);
conta.sacar(100.00);
