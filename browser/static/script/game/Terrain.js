import * as THREE from './../three/index.js';

export default (stage, world, texture) => { // https://github.com/mrdoob/three.js/blob/master/examples/webgl_geometry_minecraft.html
    // sides
    var matrix = new THREE.Matrix4();
    var pxGeometry = new THREE.PlaneGeometry(100, 100);
    pxGeometry.rotateY(Math.PI / 2);
    pxGeometry.translate(50, 0, 0);
    var nxGeometry = new THREE.PlaneGeometry(100, 100);
    nxGeometry.rotateY(- Math.PI / 2);
    nxGeometry.translate(- 50, 0, 0);
    var pyGeometry = new THREE.PlaneGeometry(100, 100);
    pyGeometry.rotateX(- Math.PI / 2);
    pyGeometry.translate(0, 50, 0);
    var pzGeometry = new THREE.PlaneGeometry(100, 100);
    pzGeometry.translate(0, 0, 50);
    var nzGeometry = new THREE.PlaneGeometry(100, 100);
    nzGeometry.rotateY(Math.PI);
    nzGeometry.translate(0, 0, -50);

    var geometry = new THREE.Geometry();
    for (var z = 0; z < world.depth; z++) {
        for (var x = 0; x < world.width; x++) {
            var h = world.getY(x, z);
            matrix.makeTranslation(
                x * 100 - world.halfWidth * 100,
                h * 100,
                z * 100 - world.halfDepth * 100
            );
            var px = world.getY(x + 1, z);
            var nx = world.getY(x - 1, z);
            var pz = world.getY(x, z + 1);
            var nz = world.getY(x, z - 1);
            const fakeRand = ((x * z * h * 22695477) | 0) % 25 - 20;
            geometry.merge(pyGeometry, matrix, 1 + Math.max(0, fakeRand));
            if ((px !== h && px !== h + 1) || x === 0) {
                geometry.merge(pxGeometry, matrix, 6);
            }
            if ((nx !== h && nx !== h + 1) || x === world.width - 1) {
                geometry.merge(nxGeometry, matrix, 6);
            }
            if ((pz !== h && pz !== h + 1) || z === world.depth - 1) {
                geometry.merge(pzGeometry, matrix, 6);
            }
            if ((nz !== h && nz !== h + 1) || z === 0) {
                geometry.merge(nzGeometry, matrix, 6);
            }
        }
    }
    console.log(geometry)
    var mesh = new THREE.Mesh(geometry, [
        new THREE.MeshLambertMaterial({ map: texture.block.dirt }),
        new THREE.MeshLambertMaterial({ map: texture.block.grass_00 }),
        new THREE.MeshLambertMaterial({ map: texture.block.grass_01 }),
        new THREE.MeshLambertMaterial({ map: texture.block.grass_02 }),
        new THREE.MeshLambertMaterial({ map: texture.block.grass_03 }),
        new THREE.MeshLambertMaterial({ map: texture.block.grass_04 }),
        new THREE.MeshLambertMaterial({ map: texture.block.grass_side }),
    ]);
    stage.add(mesh);
    var ambientLight = new THREE.AmbientLight(0xcccccc, 1);
    stage.add(ambientLight);
    var directionalLight = new THREE.DirectionalLight(0xffffff, 1);
    directionalLight.position.set(1, 1, 0.5).normalize();
    stage.add(directionalLight);
};
