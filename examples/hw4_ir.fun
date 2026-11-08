abs_and_count(value: int)
returns result: int
{
    if (value >= 0) {
        result = value + 0;
    } else {
        result = 0 - value;
    }

    while (result > 1) {
        result = result - 1;
    }
}

calculate(a: int, b: int)
returns result: int
{
    result = (a + b) * 1 + 0;
}
