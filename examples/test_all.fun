positive(x: int) => x > 0

foo(a: int, b: int)
requires a > 0
returns result: int
ensures result > 0
{
    result = a + b * 2;
    assert true;
    assume result > 0;
}

bar(a: int)
returns result: int
{
    if (a > 0) {
        result = a;
    } else {
        result = -a;
    }

    while (result > 0) {
        assert true;
        result = result - 1;
    }
}
