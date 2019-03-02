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

Need to call GLTF anim and use their Update method. In loadScene
1. Remove Prevload
2. Load the anims from .gltf json file, and fill GltfLoader.anim 

TODO:
1. Character translate, animate and generate properly.
  - Have separate body and head texture
  - Head and body remain same
  - Differences are Hair, body colour that determine Father, mom, son daughter.
  *3 points above doesnt work cuz g3n gltf doesnt support multiple multiple primitive rigging.

  - Fall back to single primitive and skin everything decide on 1 single collaged image.
  - I'll have to collage image of shirt pattern with facial image.
  - Use that collaged image as one material for 1 model.
  - Each character has own 1 .gltf and 1 picture (material)
  - So **Generate character by using right combination of .gltf and picture**
2. Model the stage.
3. Insert Skybox.

### Aim
The stage contain, ground, livestocks, and trees and grasses import all from Blender as one decoder.

Generate the character according to the choice of user, got 4 options:

1. Dad
2. Mom
3. Daughter
4. Son