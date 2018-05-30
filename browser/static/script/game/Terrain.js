import * as THREE from './../../lib/three.js/three.module.js';

export default (stage, world, texture) => { // https://github.com/mrdoob/three.js/blob/master/examples/webgl_geometry_minecraft.html
    // sides
    var matrix = new THREE.Matrix4();
    var pxGeometry = new THREE.PlaneBufferGeometry(100, 100);
    pxGeometry.attributes.uv.array[1] = 0.5;
    pxGeometry.attributes.uv.array[3] = 0.5;
    pxGeometry.rotateY(Math.PI / 2);
    pxGeometry.translate(50, 0, 0);
    var nxGeometry = new THREE.PlaneBufferGeometry(100, 100);
    nxGeometry.attributes.uv.array[1] = 0.5;
    nxGeometry.attributes.uv.array[3] = 0.5;
    nxGeometry.rotateY(- Math.PI / 2);
    nxGeometry.translate(- 50, 0, 0);
    var pyGeometry = new THREE.PlaneBufferGeometry(100, 100);
    pyGeometry.attributes.uv.array[5] = 0.5;
    pyGeometry.attributes.uv.array[7] = 0.5;
    pyGeometry.rotateX(- Math.PI / 2);
    pyGeometry.translate(0, 50, 0);
    var pzGeometry = new THREE.PlaneBufferGeometry(100, 100);
    pzGeometry.attributes.uv.array[1] = 0.5;
    pzGeometry.attributes.uv.array[3] = 0.5;
    pzGeometry.translate(0, 0, 50);
    var nzGeometry = new THREE.PlaneBufferGeometry(100, 100);
    nzGeometry.attributes.uv.array[1] = 0.5;
    nzGeometry.attributes.uv.array[3] = 0.5;
    nzGeometry.rotateY(Math.PI);
    nzGeometry.translate(0, 0, -50);

    // BufferGeometry cannot be merged yet.
    var tmpGeometry = new THREE.Geometry();
    var pxTmpGeometry = new THREE.Geometry().fromBufferGeometry(pxGeometry);
    var nxTmpGeometry = new THREE.Geometry().fromBufferGeometry(nxGeometry);
    var pyTmpGeometry = new THREE.Geometry().fromBufferGeometry(pyGeometry);
    var pzTmpGeometry = new THREE.Geometry().fromBufferGeometry(pzGeometry);
    var nzTmpGeometry = new THREE.Geometry().fromBufferGeometry(nzGeometry);
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
            tmpGeometry.merge(pyTmpGeometry, matrix);
            if ((px !== h && px !== h + 1) || x === 0) {
                tmpGeometry.merge(pxTmpGeometry, matrix);
            }
            if ((nx !== h && nx !== h + 1) || x === world.width - 1) {
                tmpGeometry.merge(nxTmpGeometry, matrix);
            }
            if ((pz !== h && pz !== h + 1) || z === world.depth - 1) {
                tmpGeometry.merge(pzTmpGeometry, matrix);
            }
            if ((nz !== h && nz !== h + 1) || z === 0) {
                tmpGeometry.merge(nzTmpGeometry, matrix);
            }
        }
    }
    var geometry = new THREE.BufferGeometry().fromGeometry(tmpGeometry);
    geometry.computeBoundingSphere();

    var mesh = new THREE.Mesh(geometry, new THREE.MeshLambertMaterial({ map: texture.grass }));
    stage.add(mesh);
    var ambientLight = new THREE.AmbientLight(0xcccccc);
    stage.add(ambientLight);
    var directionalLight = new THREE.DirectionalLight(0xffffff, 2);
    directionalLight.position.set(1, 1, 0.5).normalize();
    stage.add(directionalLight);
};
