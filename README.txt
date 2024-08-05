vdd (var dump differ) diffs two dumps from uefivardump (https://github.com/ambeloe/uefivardump) and outputs json to stdout describing the differences in the variable contents.

example usage:
  vdd -f DellMonotonicCounter dump1.json dump2.json
output:
  modified: ["DellMonotonicCounter"].Data[3] = 0xfe | a[0xff, 0xfe]b
  modified: ["DellMonotonicCounter"].Data[4] = 0xc0 | a[0xe0, 0xc0]b
