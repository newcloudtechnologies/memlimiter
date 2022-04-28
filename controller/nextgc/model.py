#!/usr/bin/env python
import math

import pandas as pd
import numpy as np

pd.set_option('display.float_format', lambda x: '%.3f' % x)


def main():
    x = [i/100 for i in range(0, 90, 10)]
    x.extend([i/100 for i in range(90, 101)])
    df = pd.DataFrame({'m_usage': x})

    danger = 0.9
    mask = df['m_usage'] < danger

    df['p'] = df['m_usage'].apply(lambda v: math.nan if v == 1 else 1/(1-v))
    df['d'] = 0
    df['sum'] = df['p'] + df['d']
    df.loc[mask, 'sum'] = 0

    gogc_default = 100
    df['GOGC'] = df['sum'].apply(lambda v: gogc_default if v == 0 else gogc_default - v)


    df['Throttling'] = df['sum'] / 100

    print(df)

if __name__ == "__main__":
    main()