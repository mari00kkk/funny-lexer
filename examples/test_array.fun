process(a: int[])
returns result: int
{
    result = a[0];
    a[1] = result + 1;
    assert result >= 0;
}
