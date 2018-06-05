import * as THREE from './three/index.js';

import * as websocket from './websocket.js';

const container = document.querySelector('#threejs');
const config = {
    get aspect() { return container.clientWidth / container.clientHeight; },
    get viewSize() { return 2; },
    get colors() {
        return [
            0xffffff, // 玩家坐标点
            0x000000, // 中心样本点
            0xff0000, 0xff9900, 0xffff00, 0x00ff00,
            0x00ffff, 0x0000ff, 0x9900ff, 0xff00ff,
        ];
    },
};

const [camera, cameraMini] = ((fov, near, far) => {
    const camera = new THREE.PerspectiveCamera(fov, config.aspect);
    camera.near = near;
    camera.far = far
    camera.position.z = config.viewSize;

    const cameraMini = new THREE.OrthographicCamera();
    cameraMini.near = near;
    cameraMini.far = far
    cameraMini.position.z = config.viewSize;

    return [camera, cameraMini];
})(60, 0.01 * config.viewSize, 10 * config.viewSize);
const [scene, sceneMini] = (() => {
    const scene = new THREE.Scene();
    scene.add(new THREE.AmbientLight(0xffffff, 0.1));
    scene.background = new THREE.Color(0x333333);
    const directionalLight = new THREE.DirectionalLight(0xffffff, 1);
    scene.add(directionalLight);
    directionalLight.position.set(-1, 1, 1);

    const sceneMini = new THREE.Scene();
    sceneMini.background = new THREE.Color(0x666666);

    return [scene, sceneMini];
})();
const [renderer, rendererMini] = (() => {
    const renderer = new THREE.WebGLRenderer({ antialias: true });
    container.appendChild(renderer.domElement);
    const rendererMini = new THREE.WebGLRenderer();
    container.appendChild(rendererMini.domElement);
    rendererMini.domElement.style.position = 'absolute';
    rendererMini.domElement.style.top = 0;
    rendererMini.domElement.style.right = 0;
    rendererMini.domElement.style.padding = '0.5em';
    const animate = () => {
        requestAnimationFrame(animate);
        renderer.render(scene, camera);
        rendererMini.render(sceneMini, cameraMini);
    };
    animate();
    return [renderer, rendererMini];
})();

(() => {
    const resize = () => {
        camera.aspect = config.aspect;
        camera.updateProjectionMatrix();
        renderer.setSize(container.clientWidth, container.clientHeight);
        cameraMini.left = - config.aspect * config.viewSize / 2;
        cameraMini.right = config.aspect * config.viewSize / 2;
        cameraMini.bottom = - config.viewSize / 2;
        cameraMini.top = config.viewSize / 2;
        cameraMini.updateProjectionMatrix();
        rendererMini.setSize(container.clientWidth / 4, container.clientHeight / 4);
    };
    window.addEventListener('resize', resize, false);
    resize();
})();

((radius) => {
    // draw sphere
    scene.add(new THREE.Mesh(
        new THREE.SphereGeometry(radius, 15, 15),
        new THREE.MeshLambertMaterial({ wireframe: true }),
    ));

    // draw helper
    const helper = new THREE.Group();
    scene.add(helper);
    const AxesHelper = new THREE.AxesHelper(radius * 1.5);
    helper.add(AxesHelper);
    AxesHelper.position.set(0, 0, 0);

    // draw points
    const points = new THREE.Points(
        new THREE.BufferGeometry(),
        new THREE.PointsMaterial({ size: config.viewSize / 200, vertexColors: true }),
    );
    const pointsMini = new THREE.Points(
        new THREE.BufferGeometry(),
        new THREE.PointsMaterial({ size: config.viewSize / 20, vertexColors: true }),
    );
    const colors = new THREE.Float32BufferAttribute(
        config.colors.reduce((cs, c) => [...cs, ...new THREE.Color(c).toArray()], []),
        3,
    );
    points.geometry.addAttribute('color', colors);
    pointsMini.geometry.addAttribute('color', colors);

    websocket.on('/api/sphere-map')(promise => promise
        .then(data => {
            points.geometry.addAttribute('position', new THREE.Float32BufferAttribute(
                data.points.reduce((ps, p) => [...ps, ...p], []), 3,
            ));
            scene.add(points);
            const projections = data.projections.reduce((ps, p) => [...ps, ...p.map(v => 10 * v), 0], []);
            const scaleRatio = config.viewSize * 0.8 / (Math.max(...projections) - Math.min(...projections));
            pointsMini.geometry.addAttribute('position', new THREE.Float32BufferAttribute(
                projections.map(v => v * scaleRatio), 3,
            ));
            sceneMini.add(pointsMini);
        })
    );
    websocket.send('/api/sphere-map');
})(1);

// controller
(() => {
    const verticalVector = new THREE.Vector3(1, 0, 0);
    const horizonVector = new THREE.Vector3(0, 1, 0);
    const parallelVector = new THREE.Vector3(0, 0, 1);
    const vertical = (angle) => { scene.rotateOnAxis(verticalVector, angle); }
    const horizon = (angle) => { scene.rotateOnAxis(horizonVector, angle); }
    const parallel = (angle) => { scene.rotateOnAxis(parallelVector, angle); }
    const reset = () => {
        scene.rotation.set(0, 0, 0);
        vertical(Math.PI / 4);
        horizon(-Math.PI / 4);
        scene.updateMatrix();
        camera.position.z = config.viewSize;
    }
    reset();
    const rotateSpeed = 0.1;
    const scaleSpeed = config.viewSize / 10;
    const controls = {
        /** View control */
        "83": reset, // S
        "87": () => { vertical(-rotateSpeed); }, // W
        "88": () => { vertical(+rotateSpeed); }, // X
        "65": () => { horizon(-rotateSpeed); },  // A
        "68": () => { horizon(+rotateSpeed); },  // D
        "69": () => { parallel(-rotateSpeed); }, // E
        "81": () => { parallel(+rotateSpeed); }, // Q
        "90": () => { camera.position.z += scaleSpeed; },                  // Z
        "67": () => { camera.position.z -= scaleSpeed; },                  // X
    };
    addEventListener("keydown", event => {
        event.stopPropagation();
        // console.log(event.keyCode);
        controls[event.keyCode] && controls[event.keyCode]();
    });
})();