type Duration = number;

const unitMap: { [key: string]: number } = {
  ns: 1,
  us: 1000,
  µs: 1000,
  ms: 1000000,
  s: 1000000000,
  m: 60000000000,
  h: 3600000000000,
};

export function parseDuration(s: string): Duration {
  const orig = s;
  let d = 0;
  let neg = false;

  // Consume [-+]?
  if (s !== '') {
    const c = s[0];
    if (c === '-' || c === '+') {
      neg = c === '-';
      s = s.slice(1);
    }
  }

  // Special case: if all that is left is "0", this is zero.
  if (s === '0') {
    return 0;
  }
  if (s === '') {
    throw new Error(`time: invalid duration "${orig}"`);
  }

  while (s !== '') {
    let v = 0;
    let f = 0;
    let scale = 1;

    // The next character must be [0-9.]
    if (!/^[0-9.]/.test(s)) {
      throw new Error(`time: invalid duration "${orig}"`);
    }

    // Consume [0-9]*
    const pl = s.length;
    [v, s] = leadingInt(s);
    const pre = pl !== s.length; // whether we consumed anything before a period

    // Consume (\.[0-9]*)?
    let post = false;
    if (s !== '' && s[0] === '.') {
      s = s.slice(1);
      const pl = s.length;
      [f, scale, s] = leadingFraction(s);
      post = pl !== s.length;
    }
    if (!pre && !post) {
      // no digits (e.g. ".s" or "-.s")
      throw new Error(`time: invalid duration "${orig}"`);
    }

    // Consume unit.
    const i = s.search(/[^a-zµ]/i);
    const u = i === -1 ? s : s.slice(0, i);
    s = i === -1 ? '' : s.slice(i);
    if (u === '') {
      throw new Error(`time: missing unit in duration "${orig}"`);
    }
    const unit = unitMap[u.toLowerCase()];
    if (unit === undefined) {
      throw new Error(`time: unknown unit "${u}" in duration "${orig}"`);
    }

    // Calculate value without causing intermediate overflow
    let value = 0;
    const maxBeforeOverflow = Math.floor((Number.MAX_SAFE_INTEGER - d) / unit);

    if (v > maxBeforeOverflow) {
      throw new Error(`time: invalid duration "${orig}"`);
    }

    value = v * unit;

    if (f > 0) {
      const fraction = Math.floor(f * (unit / scale));
      if (value > Number.MAX_SAFE_INTEGER - fraction) {
        throw new Error(`time: invalid duration "${orig}"`);
      }
      value += fraction;
    }

    if (d > Number.MAX_SAFE_INTEGER - value) {
      throw new Error(`time: invalid duration "${orig}"`);
    }
    d += value;
  }

  return neg ? -d : d;
}

function leadingInt(s: string): [number, string] {
  let i = 0;
  while (i < s.length && '0' <= s[i] && s[i] <= '9') {
    i++;
  }
  if (i == 0) {
    throw new Error('time: bad [0-9]*');
  }
  return [parseInt(s.slice(0, i), 10), s.slice(i)];
}

function leadingFraction(s: string): [number, number, string] {
  let i = 0;
  let scale = 1;
  while (i < s.length && '0' <= s[i] && s[i] <= '9') {
    i++;
    scale *= 10;
  }
  return [parseInt(s.slice(0, i), 10), scale, s.slice(i)];
}
