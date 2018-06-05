import * as THREE from './three/index.js';

import Controller from './game/Controller.js';
import Terrain from './game/Terrain.js';
import texture from './game/texture.js';
import * as websocket from './websocket.js';
import * as debug from './debug.js';

const container = document.querySelector('#threejs');
const config = {
    get aspect() { return container.clientWidth / container.clientHeight; },
    get color() {
        return { sky: 0xbfd1e5 };
    },
    get sight() { return 64 },
};

const camera = ((fov, near, far) => {
    const camera = new THREE.PerspectiveCamera(fov, config.aspect);
    camera.near = near;
    camera.far = far;
    return camera;
})(60, 1, 10 * 1000);
const scene = (() => {
    const scene = new THREE.Scene();
    scene.background = new THREE.Color(config.color.sky);
    scene.fog = new THREE.FogExp2(config.color.sky, 0.05 / 1000);
    return scene;
})();
const renderer = (() => {
    const renderer = new THREE.WebGLRenderer();
    container.appendChild(renderer.domElement);
    renderer.setPixelRatio(window.devicePixelRatio);
    const controller = Controller(camera, renderer.domElement, {
        moveSpeed: 1000, jumpSpeed: 1000, lookSpeed: 300, pitchSpeed: 150,
    });
    // const stats = new Stats();
    // container.appendChild(stats.dom);
    const animate = () => {
        requestAnimationFrame(animate);
        controller.update();
        renderer.render(scene, camera);
        // stats.update();
    };
    animate();
    return renderer;
})();

(() => {
    const resize = () => {
        camera.aspect = config.aspect;
        camera.updateProjectionMatrix();
        renderer.setSize(container.clientWidth, container.clientHeight);
    };
    window.addEventListener('resize', resize, false);
    resize();
})();

(() => {
    const stage = new THREE.Group();
    scene.add(stage);
    const loadTerrain = async ({ width, depth, data }) => {
        stage.remove(...stage.children);
        const actualWidth = Math.min(config.sight, width);
        const actualDepth = Math.min(config.sight, depth);
        const world = {
            get width() { return actualWidth; },
            get halfWidth() { return actualWidth / 2; },
            get depth() { return actualDepth; },
            get halfDepth() { return actualDepth / 2; },
            getY: (x, z) => data[x + z * width] | 0,
        };
        camera.position.y = world.getY(world.halfWidth, world.halfDepth) * 100 + 100;
        Terrain(stage, world, await texture);
    };
    websocket.on('/api/world')(promise => promise.then(data => loadTerrain(data.terrain)));
    websocket.send('/api/world');

    debug.exports('terrain', async exports => {
        // const ImprovedNoise = await fetch('https://threejs.org/examples/js/ImprovedNoise.js').then(r => r.text()).then(
        //     text => eval(`(() => { ${text} return ImprovedNoise; })();`)
        // );
        // const perlin = new ImprovedNoise();
        // let [width, depth] = [128, 128];

        // Object.defineProperties(exports, {
        //     noise: { get: () => perlin.noise },
        //     width: { set: v => width = v },
        //     depth: { set: v => depth = v },
        // });
        // exports.default = (scale, amplitude, seed = new Date().getTime()) => {
        //     const size = width * depth;
        //     const data = new Array(size).fill(0);
        //     let quality = 2;
        //     for (let j = 0; j < 4; j++) {
        //         for (let i = 0; i < size; i++) {
        //             var [x, y] = [i % width, (i / width) | 0];
        //             data[i] += amplitude * perlin.noise(x / quality, y / quality, seed) * quality;
        //         }
        //         quality *= scale;
        //     }
        //     loadTerrain({ width, depth, data });
        // };

        // // exports.width = 1;
        // // exports.depth = 1024;
        // // camera.far = 1000 ** 3;
        // // scene.fog = null;
        // // camera.updateProjectionMatrix();
        // // exports.default(5, 1 / 5, 0);
    });
})();
