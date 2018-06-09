import * as THREE from './three/index.js';

import Controller from './game/Controller.js';
import Stage from './game/Stage.js';
import texture from './game/texture.js';
import * as websocket from './websocket.js';
import * as debug from './debug.js';

const container = document.querySelector('#threejs');
const config = {
    get aspect() { return container.clientWidth / container.clientHeight; },
    get color() {
        return { sky: 0xbfd1e5 };
    },
    get sight() { return 32; },
};

const camera = ((fov, near, far) => {
    const camera = new THREE.PerspectiveCamera(fov, config.aspect);
    camera.near = near;
    camera.far = far;
    return camera;
})(70, 1, 5 * 1000);
const scene = (() => {
    const scene = new THREE.Scene();
    scene.background = new THREE.Color(config.color.sky);
    scene.fog = new THREE.FogExp2(config.color.sky, 0.05 / 1000);
    return scene;
})();

const [renderer, requestRender] = (() => {
    const renderer = new THREE.WebGLRenderer();
    container.appendChild(renderer.domElement);
    renderer.setPixelRatio(window.devicePixelRatio);
    // const stats = new Stats();
    // container.appendChild(stats.dom);
    const renderQueue = [];
    const animate = () => {
        requestAnimationFrame(animate);
        renderQueue.forEach(action => action());
        renderer.render(scene, camera);
        // stats.update();
    };
    animate();
    return [renderer, action => renderQueue.push(action)];
})();

(() => {
    const stage = Stage(camera, scene, texture, {
        sight: config.sight, requestRender,
    });
    requestRender(() => stage.update());
    const controller = Controller(camera, renderer.domElement, {
        moveSpeed: 1000, jumpSpeed: 1000, lookSpeed: 300, pitchSpeed: 150,
        handler: stage.handler,
    });
    requestRender(() => controller.update());

    websocket.on('/api/world')(promise => promise.then(stage.loadTerrain));
    websocket.send('/api/world')

    debug.exports('terrain', async exports => {
        const ImprovedNoise = await fetch('https://threejs.org/examples/js/ImprovedNoise.js').then(r => r.text()).then(
            text => eval(`(() => { ${text} return ImprovedNoise; })();`)
        );
        const perlin = new ImprovedNoise();
        let [width, depth] = [128, 128];

        Object.defineProperties(exports, {
            noise: { get: () => perlin.noise },
            width: { set: v => width = v },
            depth: { set: v => depth = v },
        });
        exports.default = (scale, amplitude, seed = new Date().getTime()) => {
            const size = width * depth;
            const heights = new Array(size).fill(0);
            let quality = 2;
            for (let j = 0; j < 4; j++) {
                for (let i = 0; i < size; i++) {
                    var [x, y] = [i % width, (i / width) | 0];
                    heights[i] += amplitude * perlin.noise(x / quality, y / quality, seed) * quality;
                }
                quality *= scale;
            }
            stage.clear();
            stage.loadTerrain({
                x: -width / 2, z: -depth / 2, width, depth,
                data: heights.map(h => [h | 0, 1]),
            });
        };

        // exports.width = 1;
        // exports.depth = 1024;
        // camera.far = 1000 ** 3;
        // scene.fog = null;
        // camera.updateProjectionMatrix();
        // exports.default(5, 1 / 5, 0);
    });

    debug.exports('command', exports => {
        exports.shiftChunk = id => websocket.send('/api/chunk/shift', { id }).then(() => {
            stage.clear();
            websocket.send('/api/world');
        });
    })
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
