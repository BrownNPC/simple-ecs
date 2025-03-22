# Simple-ECS
dead-simple library for writing
game systems in Go

### Simple-ECS Features:
- Easy syntax / api
- Good perfomance
- Easy to understand code (300 lines*)
- Low level (implement what you need)

### Get started
 - I am a beginner
 - I know what ECS is



### What is ECS? (and why you should use it)

ECS stands for entity component system.

some games follow Object Oriented Design,
where everything in the game is an object (a class)
and every entity ("living things") uses inheritence
to build up different types of gameplay.

For example, you may have a class called
"Entity" which has a position and a sprite.

Then you have a class called **Actor** that inherits from
"Entity" and this actor has methods for speaking, moving etc.

Also inheriting from the Entity, you can have a
**Tree** class.

From Actor, we make **Enemy** that has methods
for enemy behaviour.

Now, what if we want to make an Evil Tree?







### Motivation + Opinion:
  The other ECS libraries seem
  to focus on having the best
  possible performance,
  sometimes sacrificing a
  simpler syntax. They also provide features
  I dont need.
  Some devs put
  restrictions on the use of
  runtime reflection for negligible
  performance gains. And these libraries had
  many ways to do
  the same thing. (eg. Arche has 2 apis)

  This is just my opinion but most
  games that are made using Go should
  not care about microseconds worth
  of performance gains. As the main reason
  to pick Go over C++, C#, Java or Rust is
  because of Go's simplicity.
  
  Also no hate or anything of that sort
  is intended towards any developer's work.
  Everyone has their own reasons for writing
  their own code. I am not claiming that
  this is the best ECS for Go. I am only claiming
  that it has a simple API,
  but that could be subjective.

### Acknowledgements
  Donburi is another library that
  implements ECS with a simple API.
  But in my opinion this library is
  simpler.
