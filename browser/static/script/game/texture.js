import * as THREE from './../three/index.js';
import * as block from './../../asset/texture/block.js';
import * as debug from './../debug.js';

const dir = './static/asset/texture';
const asyncCollector = async (previous, current) => ({ ...await previous, ...await current });
const loader = new THREE.TextureLoader();

/** @type {PromiseLike<Object.<string, SpriteInfo>>} */
const packer = Object.entries({ block: block.default }).map(
    async ([name, sprites]) => {
        /** @type {THREE.Texture} */
        const packed = await new Promise((resolve, reject) => loader.load(`${dir}/${name}.png`, resolve, undefined, reject));
        const textures = await Object.entries(sprites).map(
            async ([name, { rect: { x, y, width, height } }]) => {
                const texture = packed.clone();
                texture.magFilter = THREE.NearestFilter;
                texture.minFilter = THREE.LinearMipMapLinearFilter;
                texture.repeat.set(width / packed.image.width, height / packed.image.height);
                texture.offset.x = (x) / packed.image.width;
                texture.offset.y = 1 - (height / packed.image.height) - (y / packed.image.height);
                texture.needsUpdate = true;
                // return { [name]: loader.load('https://threejs.org/examples/textures/minecraft/atlas.png') };
                return { [name]: texture };
            }
        ).reduce(asyncCollector, Promise.resolve({}));
        return {
            [name]: new Proxy(textures, {
                get: (target, key, _receiver) => {
                    const id = key.toLowerCase();
                    if (textures.hasOwnProperty(id)) {
                        return target[id];
                    }
                    console.error(`missing texture: ${id}.`);
                    return target[404];
                },
                has: (target, key, _receiver) => textures.hasOwnProperty(key),
            }),
        };
    }
).reduce(asyncCollector, Promise.resolve({}));

export default packer;
debug.exports('texture', async exports => exports.default = await packer);
