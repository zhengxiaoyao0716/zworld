import * as THREE from './../lib/three.js/three.module.js';

import Controller from './game/Controller.js';
import Terrain from './game/Terrain.js';
import * as websocket from './websocket.js';

const container = document.querySelector('#threejs');
const config = {
    get aspect() { return container.clientWidth / container.clientHeight; },
    get texture() {
        const loader = new THREE.TextureLoader();
        const grass = loader.load('https://threejs.org/examples/textures/minecraft/atlas.png');
        grass.magFilter = THREE.NearestFilter;
        grass.minFilter = THREE.LinearMipMapLinearFilter;
        return { grass };
    },
    get color() {
        return { sky: 0xbfd1e5 };
    },
};

const camera = ((fov, near, far) => {
    const camera = new THREE.PerspectiveCamera(fov, config.aspect);
    camera.near = near;
    camera.far = far;
    return camera;
})(60, 1, 10000);
const scene = (() => {
    const scene = new THREE.Scene();
    scene.background = new THREE.Color(config.color.sky);
    scene.fog = new THREE.FogExp2(config.color.sky);
    return scene;
})();

const renderer = (() => {
    const renderer = new THREE.WebGLRenderer();
    container.appendChild(renderer.domElement);
    renderer.setPixelRatio(window.devicePixelRatio);
    const controller = Controller(camera, renderer.domElement, {
        moveSpeed: 1000, lookSpeed: 5,
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

const initWorld = (length, higth) => {
    const width = 128, depth = 128;
    const data = ((width, height) => {
        var data = [], perlin = new ImprovedNoise(),
            size = width * height, quality = 2, z = Math.random() * 100;
        for (var j = 0; j < 4; j++) {
            if (j === 0) for (var i = 0; i < size; i++) data[i] = 0;
            for (var i = 0; i < size; i++) {
                var x = i % width, y = (i / width) | 0;
                data[i] += perlin.noise(x / quality, y / quality, z) * quality;
            }
            quality *= length;
        }
        return data;
    })(width, depth);
    return {
        get width() { return width; }, get halfWidth() { return width / 2; },
        get depth() { return depth; }, get halfDepth() { return depth / 2; },
        getY: (x, z) => (data[x + z * width] * higth) | 0,
    };
};
(() => {
    const stage = new THREE.Group();
    scene.add(stage);
    const loadTerrain = (length, higth, { _width, _depth, _data }) => {
        stage.remove(...stage.children);
        const world = initWorld(length, higth, );
        const { halfWidth, halfDepth, getY } = world;
        camera.position.y = world.getY(world.halfWidth, world.halfDepth) * 100 + 100;
        Terrain(stage, world, config.texture);
    };
    window.loadTerrain = loadTerrain;
    websocket.on('/api/world')(promise => promise.then(data => loadTerrain(5, 1 / 5, data)));
    websocket.send('/api/world');
})();
