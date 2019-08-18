# golsd
A small experiment learning how to use [GopherJS](https://github.com/gopherjs/gopherjs) and [beep](https://github.com/faiface/beep), playing off of a large experiment by the [federal government](https://youtu.be/Esx_uvr3Rf4).

## What's going on?
I really enjoy archived footage, and I really enjoy the musicality of human voice. These interests intersect in this project, which features speech clippings from a documentary about the Spring Grove Experiment aligned with synth percussion. These clippings are played in a random order, ad infinitum, with sinusoidal mixing between the speech and the percussion.

I use golang a lot at work, and was interested in its use in the audio space, which led me to the great library [beep](https://github.com/faiface/beep). The web is great for sharing projects, which is why I was very excited to see that beep and its underlying audio engine [oto](https://github.com/hajimehoshi/oto) play well with [GopherJS](https://github.com/gopherjs/gopherjs. It was surprisingly straightforward to have my go code running in and interacting with the browser!

## Reference
Neher, Jack (1967). "LSD: The Spring Grove Experiment (54 minutes, black and white. Produced by CBS)". Psychiatric Services. 18 (5): 157–a–157. doi:10.1176/ps.18.5.157-a