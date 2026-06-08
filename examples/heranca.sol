rise ContaBancaria {

    private float saldo;
    private string titular;

    glow(string titular, float saldoInicial) {
        this.titular = titular;
        this.saldo   = saldoInicial;
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
}

rise ContaEspecial radiate ContaBancaria {

    private float limiteCredito;

    glow(string titular, float saldoInicial, float limiteCredito) {
        radiate.glow(titular, saldoInicial);
        this.limiteCredito = limiteCredito;
    }

    public ray sacar(float valor) {
        var disponivel float = this.getSaldo() + this.limiteCredito;
        if (valor > disponivel) {
            flare "Total limit exceeded";
        }
        radiate.sacar(valor);
    }

    public ray getLimite() {
        emit this.limiteCredito;
    }
}

var especial ContaEspecial = new ContaEspecial("Carlos", 500.00, 300.00);
especial.sacar(50.00);
