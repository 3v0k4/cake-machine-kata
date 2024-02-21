# Cake Machine Kata (concurrent/parallel programming)

Write a console program to simulate a cake manufacturing factory.

Make as many cakes as possible. But don't worry about the ingredients, they are infinite.

A cake is ready after 3 stages:
1. Preparation: random duration between 5 and 8 seconds
1. Cooking: 10 seconds
1. Packaging: 2 seconds

Each stage has a different amount of workers:
- Preparation: 3 cakes at the same time
- Cooking: 5 cakes at the same time
- Packing: 2 cakes at the same time

Every minute, a report displays the number of cakes completed and the number of cakes at each stage of preparation.
