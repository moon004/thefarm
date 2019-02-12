A client (louis) project that is under construction

### Powered by Blender, Golang combined

### Architecture
The Farm is the main parent object. It has the following vital things:

>stageScene **Node* 

***Node is an object in g3n***. **Must add to stageScene in order to get rendered.**

>stage *Stage

**Stage is the platform that characters live on, it includes all other props like trees and livestocks.**

>charNode *Node

**The character node, generated from input (facial pictures)**

