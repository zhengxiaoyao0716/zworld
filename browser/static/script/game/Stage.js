import * as THREE from './../three/index.js';
import * as websocket from './../websocket.js';
import * as debug from './../debug.js';
// import Physijs from './../Physijs/physi.js';

const geometries = (() => {
    const pxGeometry = new THREE.PlaneBufferGeometry(100, 100);
    pxGeometry.rotateY(Math.PI / 2);
    pxGeometry.translate(50, 0, 0);
    const nxGeometry = new THREE.PlaneBufferGeometry(100, 100);
    nxGeometry.rotateY(- Math.PI / 2);
    nxGeometry.translate(- 50, 0, 0);
    const pyGeometry = new THREE.PlaneBufferGeometry(100, 100);
    pyGeometry.rotateX(- Math.PI / 2);
    pyGeometry.translate(0, 50, 0);
    const nyGeometry = new THREE.PlaneBufferGeometry(100, 100);
    nyGeometry.rotateX(Math.PI / 2);
    nyGeometry.translate(0, -50, 0);
    const pzGeometry = new THREE.PlaneBufferGeometry(100, 100);
    pzGeometry.translate(0, 0, 50);
    const nzGeometry = new THREE.PlaneBufferGeometry(100, 100);
    nzGeometry.rotateY(Math.PI);
    nzGeometry.translate(0, 0, -50);
    return [pxGeometry, nxGeometry, pyGeometry, nyGeometry, pzGeometry, nzGeometry];
})();

const mesheIds = {
    block: {
        names: ['dirt', 'grass', 'rock'],
        nameOf(id) { return this.names[id]; },
        idOf(name) { return this.names.indexOf(name); }
    },
};

const Meshes = texture => {
    const cache = {
        block: {},
    };
    const createBlock = (name, ...textureIds) => {
        const materials = textureIds.map(id => new THREE.MeshLambertMaterial({ map: texture.block[id] }));
        return {
            /**
             * @param {number[]|{ x: number, y: number, z: number }} position .
             * @param {number[]} faceIds .
             */
            create: (position, ...faceIds) => {
                const [x, y, z] = position instanceof Array ? position : [position.x, position.y, position.z];
                const block = new THREE.Group();
                block.name = name;
                block.position.set(x, y, z);
                const addMesh = (faceId, materialId) => {
                    // const mesh = new Physijs.Mesh(geometries[faceId], materials[materialId]);
                    const mesh = new THREE.Mesh(geometries[faceId], materials[materialId]);
                    block.add(mesh);
                    mesh.position.set(x, y, z);
                    mesh.userData.faceId = faceId;
                };
                (faceIds.length == 0 ? geometries.map((_, i) => i) : faceIds).forEach(
                    faceId => Number.isInteger(faceId)
                        ? addMesh(faceId, faceId)
                        : addMesh(faceId.geometry, faceId.material)
                );
                return block;
            }
        };
    };
    return {
        block: new Proxy(
            {
                grass: [
                    'grass_side', 'grass_side', 'grass_00', 'dirt', 'grass_side', 'grass_side', // main
                    'grass_01', 'grass_02', 'grass_03', 'grass_04', // extra
                ],
            },
            {
                get: (target, key, _receiver) => {
                    if (cache.block.hasOwnProperty(key)) {
                        return cache.block[key];
                    }
                    const block = target.hasOwnProperty(key)
                        ? createBlock(key, ...target[key])
                        : createBlock(
                            key,
                            `${key}_px` in texture ? `${key}_px` : key,
                            `${key}_nx` in texture ? `${key}_nx` : key,
                            `${key}_py` in texture ? `${key}_py` : key,
                            `${key}_ny` in texture ? `${key}_ny` : key,
                            `${key}_pz` in texture ? `${key}_pz` : key,
                            `${key}_nz` in texture ? `${key}_nz` : key,
                        );
                    cache.block[key] = block;
                    return block;
                },
            },
        ),
    };
};

const attachGroup = (parent, group) => {
    const linkGroup = object => {
        if (!object.userData.group) {
            object.userData.group = group;
            return object;
        }
        if (object.userData.group != group) {
            linkGroup(object.userData.group);
        }
        return object;
    }
    group.userData.group = parent;
    group.userData.children = group.userData.children || group.children.map(linkGroup);
    if (!parent.userData.group) {
        parent.add(...group.userData.children);
        return;
    }
    parent.userData.children.push(...group.userData.children);
    const addToRoot = object => object.userData.group
        ? addToRoot(object.userData.group) : object.add(...group.userData.children);
    addToRoot(parent.userData.group);
};
const detachGroup = group => {
    const freeGroup = object => {
        if (!object.userData.group) {
            object.remove(...group.userData.children);
            return;
        }
        freeGroup(object.userData.group);
    }
    const parent = group.userData.group;
    freeGroup(parent);
    if (parent.userData.children) {
        const children = new Set(group.userData.children);
        parent.userData.children = parent.userData.children.filter(object => !children.has(object));
    }
    group.add(...group.userData.children);
    group.userData.group = null;
    group.userData.children = null;
    return parent;
};

export default (camera, scene, texture, { sight, requestRender }) => {
    const ambientLight = new THREE.AmbientLight(0xcccccc, 1);
    scene.add(ambientLight);
    const directionalLight = new THREE.DirectionalLight(0xffffff, 0.5);
    directionalLight.position.set(1, 1, 0.5).normalize();
    scene.add(directionalLight);
    // const stage = new Physijs.Scene();
    // requestRender(() => stage.simulate());
    const stage = new THREE.Group();
    scene.add(stage);
    debug.exports('stage', exports => {
        exports.stage = stage;
    });

    const meshes = texture.then(Meshes);

    const handler = (() => {
        const raycaster = new THREE.Raycaster();
        const intersectObjects = () => {
            raycaster.setFromCamera({ x: 0, y: 0 }, camera);
            return raycaster.intersectObjects(stage.children);
        };
        const flatObjects = (objects, object3d) => [...objects, ...(
            object3d instanceof THREE.Group ? object3d.children.reduce(flatObjects, []) : [object3d]
        )];
        const sendBuild = websocket.batchSender(
            '/api/world/build',
            (batch, { block }) => ({ block: [...batch.block, ...block] }),
        );
        return {
            actLeft: timer => {
                const intersect = intersectObjects();
                if (intersect.length == 0) {
                    return;
                }
                const { userData: { group: block } } = intersect[0].object;
                if (block.userData.durable == null) {
                    block.userData.durable = 3;
                    return;
                }
                if (block.userData.durable > timer) {
                    return;
                }
                const terrain = detachGroup(block);
                const position = [block.position.x, block.position.y - 100, block.position.z];
                meshes.then(mesh => mesh.block.dirt.create(position, 2))
                    .then(dirt => attachGroup(terrain, dirt));
                sendBuild({
                    block: [{
                        x: position[0] / 100, z: position[2] / 100,
                        h: position[1] / 100, id: mesheIds.block.idOf('dirt'),
                    }],
                });
                return true; // 重新计时
            },
            actRight: () => {
                const intersect = intersectObjects();
                if (intersect.length == 0) {
                    return;
                }
                const { face, object: { userData: { group: block } } } = intersect[0];
                const position = block.position.add(face.normal.multiplyScalar(100));
                meshes.then(mesh => mesh.block.rock.create(position))
                    .then(rock => attachGroup(block.userData.group, rock));
                sendBuild({
                    block: [{
                        x: position.x / 100, z: position.z / 100,
                        h: position.y / 100, id: mesheIds.block.idOf('rock'),
                    }],
                });
            },
        };
    })();

    const terrains = (() => {
        const terrains = {};
        const actived = new Set();
        const nowPosIndex = { xi: 0, zi: 0 };
        const fixPlayerPosition = ({ getY }) => {
            const [h,] = getY(camera.position.x, camera.position.z);
            h != null && camera.position.z <= h * 100 && (camera.position.y = (h + 2) * 100);
        }
        const update = async terrainData => {
            const { xi, zi, getY, terrain: cached } = terrainData;
            const terrain = cached || await createTerrain(`stage_terrain_${xi}_${zi}`, {
                x: xi * sight, z: zi * sight, sight, getY,
            }, meshes);
            terrainData.terrain = terrain;
            attachGroup(stage, terrain);
            actived.add({ xi, zi, terrain });
        };
        const validateChunk = terrainData => terrainData.getY != null;
        return {
            push: ({ x: startX, z: startZ, width, depth, data, ...extra }) => {
                const [startXi, startZi] = [Math.floor(startX / sight), Math.floor(startZ / sight)];
                if (!data) {
                    if (extra.out) { // 区块越界
                        terrains[[startXi, startZi]] = extra;
                    }
                    return;
                }
                const getY = (x, z) => {
                    if (x < startX || z < startZ || x >= startX + width || z >= startZ + depth) {
                        return [];
                    }
                    return data[x - startX + (z - startZ) * width];
                };
                const [nxi, nzi] = [nowPosIndex.xi, nowPosIndex.zi];
                for (let xi = startXi; xi < Math.floor((startX + width) / sight); xi++) {
                    for (let zi = startZi; zi < Math.floor((startZ + depth) / sight); zi++) {
                        const terrainData = { xi, zi, getY };
                        if (Math.abs(xi - nxi) <= 1 && Math.abs(zi - nzi) <= 1) {
                            const terrain = terrains[[xi, zi]];
                            terrain && detachGroup(terrain);
                            update(terrainData);
                            fixPlayerPosition(terrainData);
                        }
                        terrains[[xi, zi]] = terrainData;
                    }
                }
                return getY;
            },
            update: () => {
                const { x, y, z } = camera.position;
                const [nxi, nzi] = [Math.floor(x / 100 / sight), Math.floor(z / 100 / sight)];
                if (nowPosIndex.xi == nxi && nowPosIndex.zi == nzi) {
                    return;
                }
                nowPosIndex.xi = nxi;
                nowPosIndex.zi = nzi;
                actived.forEach(({ xi, zi, terrain }) =>
                    (Math.abs(xi - nxi) > 1 || Math.abs(zi - nzi) > 1) && detachGroup(terrain)
                );
                actived.clear();
                Promise.resolve(terrains[[nxi, nzi]]).then(chunk => {
                    if (!validateChunk(chunk)) {
                        // TODO 当前所站立区块已经越界
                        console.log(chunk);
                        return;
                    }
                    update(chunk);
                    fixPlayerPosition(chunk);
                });
                [
                    [-1, +0], [+0, +1], [+1, +0], [+0, -1],
                    [-1, -1], [-1, +1], [+1, -1], [+1, +1],
                ].map(([dxi, dzi]) => [nxi + dxi, nzi + dzi]).map(
                    ([xi, zi]) => {
                        const chunk = terrains[[xi, zi]];
                        if (!chunk) {
                            websocket.send("/api/world", { x: xi * sight, z: zi * sight, width: sight, depth: sight });
                            return;
                        }
                        validateChunk(chunk) && update(chunk);
                    }
                );
            },
            clear: () => {
                Object.keys(terrains).forEach(key => delete terrains[key]);
                actived.clear();
                nowPosIndex.xi = 0;
                nowPosIndex.zi = 0;
            },
        };
    })();

    return {
        handler,
        loadTerrain: terrains.push,
        update: () => {
            terrains.update();
        },
        clear: () => {
            stage.remove(...stage.children);
            terrains.clear();
        },
    };
};

const createTerrain = (name, { x: startX, z: startZ, sight, getY }, meshes) => { // https://github.com/mrdoob/three.js/blob/master/examples/webgl_geometry_minecraft.html
    const terrain = new THREE.Group();
    terrain.name = name;
    const tasks = [];
    for (let z = startZ; z < startZ + sight; z++) {
        for (let x = startX; x < startX + sight; x++) {
            const [h, blockId] = getY(x, z);
            if (h == null) {
                continue;
            }
            const [px,] = getY(x + 1, z);
            const [nx,] = getY(x - 1, z);
            const [pz,] = getY(x, z + 1);
            const [nz,] = getY(x, z - 1);

            const faces = [];
            const blockName = mesheIds.block.nameOf(blockId);
            if (blockName == 'grass') {
                const fakeRand = Math.max(0, ((x * z * h * 22695477) | 0) % 25 - 20);
                faces.push({ geometry: 2, material: fakeRand > 0 ? 5 + fakeRand : 2 });
            } else {
                faces.push(2);
            }
            if ((px !== h && px !== h + 1) || x === 0) {
                faces.push(0);
            }
            if ((nx !== h && nx !== h + 1) || x === sight - 1) {
                faces.push(1);
            }
            if ((pz !== h && pz !== h + 1) || z === sight - 1) {
                faces.push(4);
            }
            if ((nz !== h && nz !== h + 1) || z === 0) {
                faces.push(5);
            }
            const task = meshes.then(({ block }) => attachGroup(
                terrain,
                block[blockName].create([x * 100, h * 100, z * 100], ...faces),
            ));
            tasks.push(task);
        }
    }
    return Promise.all(tasks).then(() => terrain);
};
