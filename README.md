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
1. Character animate and generate properly.
  - Have separate body and head texture
  - Head and body remain same
  - Differences are Hair, body colour that determine Father, mom, son daughter.
  *3 points above doesnt work cuz g3n gltf doesnt support multiple multiple primitive rigging.

  - Fall back to single primitive and skin everything decide on 1 single collaged image.
  - I'll have to collage image of shirt pattern with facial image.
  - Use that collaged image as one material for 1 model.
  - Each character has own 1 .gltf and 1 picture (material)
  - So **Generate character by using right combination of .gltf and picture(example: gombine father.png facial1.png -out /face/user1.png)**
  Reminder: the gltf is correspond to user facial picture (change to user facial pic name at gltf.go ***around line 41***)
2. Model the stage. (Done)

3. Translate characters (done)
  - Each Character must hv own node, hence, slice of tf.charNode is created
  - Each node is then assign random dest, once reach there, assign a new random dest
  - Y kept at 0, X and Z are treated as separated components to make translation easier.
  - Rotate based on Inverse Tangent (math32.Atan2())

4. Gombine Picture, and generate characters properly
  - Use gocv, use AI (SSD facedetect) to get the face rectangle and crop it and straight away gombine with the selected model.
  - Save the resulted file (image) in a specific folder.
  
5. Insert Skybox. (Blue Sky)
6. Add character selection features.

### Aim
The stage contain, ground, livestocks, and trees and grasses import all from Blender.

Generate the character according to the choice of user, got 4 options:

1. Dad
2. Mom
3. Daughter
4. Son