check(a: int, b: int)
requires not (a < 0) and b > 0 or a == b
returns result: int
{
    result = a;
    assert not (a < 0) and b > 0;
    assume a == b or a > b;
}
