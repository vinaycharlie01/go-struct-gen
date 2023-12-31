const arr = [
    {
      place: "san francisco",
      name: "jane",
    },
    {
      place: "san francisco",
      name: "jane",
    },
    {
      place: "new york",
      name: "james",
    },
  ];
  const result = arr.filter(
    (thing, index, self) =>
      index ===
      self.findIndex((t) => t.place === thing.place && t.name === thing.name)
  );
  console.log(result);