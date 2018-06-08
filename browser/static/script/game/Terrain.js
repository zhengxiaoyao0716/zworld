import * as THREE from './../three/index.js';

export default (stage, world, texture) => { // https://github.com/mrdoob/three.js/blob/master/examples/webgl_geometry_minecraft.html
    const blockGeometry = (() => {
        const geometry = new THREE.BoxGeometry(100, 100, 100);
        geometry.faces.forEach((face, i) => face.materialIndex = i / 2 | 0);
        return geometry;
    })();
    const blockMaterials = [
        new THREE.MeshLambertMaterial({ map: texture.block.grass_side }), // back
        new THREE.MeshLambertMaterial({ map: texture.block.grass_side }), // front
        new THREE.MeshLambertMaterial({ map: texture.block.grass_00 }), // top
        new THREE.MeshLambertMaterial({ map: texture.block.dirt }), // bottom
        new THREE.MeshLambertMaterial({ map: texture.block.grass_side }), // right
        new THREE.MeshLambertMaterial({ map: texture.block.grass_side }), // left
    ];
    for (var z = 0; z < world.depth; z++) {
        for (var x = 0; x < world.width; x++) {
            var h = world.getY(x, z);
            const block = new THREE.Mesh(blockGeometry, blockMaterials);
            block.position.set(
                x * 100 - world.halfWidth * 100,
                h * 100,
                z * 100 - world.halfDepth * 100
            );
            stage.add(block);
        }
    }
    var ambientLight = new THREE.AmbientLight(0xcccccc, 1);
    stage.add(ambientLight);
    var directionalLight = new THREE.DirectionalLight(0xffffff, 1);
    directionalLight.position.set(1, 1, 0.5).normalize();
    stage.add(directionalLight);
};
