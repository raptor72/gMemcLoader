# gMemcLoader
Concurrent loader files to memchached


Программа парсит и заливает в мемĸеш поминутную
выгрузĸу логов треĸера установленных приложений. Ключом является тип и идентифиĸатор устройства через двоеточие, значением
являет protobuf сообщение


Ссылĸи на tsv.gz файлы:

 - https://cloud.mail.ru/public/2hZL/Ko9s8R9TA
 - https://cloud.mail.ru/public/DzSX/oj8RxGX1A
 - https://cloud.mail.ru/public/LoDo/SfsPEzoGc

Пример содержания файлов:
```bash
zcat ./20170929000200.tsv.gz | head -n 3
gaid	e6143e026216ba6ae8a1eb3b8be86508	130.602378477	-36.0939908322	7303,8936,3797,4542,108,6678,1703,9864,36,4091,798,1909,1946,3574,6735,7542,8464,2332,628,7416,3006,7171,2134,2670,9949,3367,619,8542,1201
dvid	711ea9bcf70928202cc33038bd4263af	-49.915096804	-16.6746214727	7330,5234,4123,8075,451,1578,1075,4324,9916,1536,9303,4613,1856,5853,8008,4131,3199,2264,7750,3680,6370,8148,5726,4559,8211,9718,1869,8085,3121,4721
idfa	02b073ec7f1b2a64d3e299e0cb1a07b1	-144.976035927	-82.9647017005	5084,1834,2236,9980,5003,3085,1198,3656,5991,5837,8179,8686,9460,8998,8464,339,760,2822,3843,7091,3316,3076,3279,2561,5491,7899,4551,8325,7668,1473,1075,6355,1055,8962,6545,9125,9027,711,1252,3239,6047,5739,6119,9265,3989,347,2619,4391,3173,9636,3597,2604,4758,766,7267,2311,3782,5603,6995,568,3738,7633,9571,9635
```

Файлы будут обработаны в хронологичесĸом порядĸе. Под этим имеется в виду, что после обработĸи они
будут префиĸсированы точĸой, последовательно согласно сортировке по имени. Заливаться в memcached при этом они будут паралельно.

Для корректной работы необходимо настроить memcached для хранения требуемого количества значений, выделив соответствующее количество памяти.

Для кеширования файлов в текущей директории поместите исходник в директорию с файлами и выполните:

```bash
go run ./memcLoader.go
```

Для получения справки по опциям выполните:
```bash
go run ./memcLoader.go --help
```

Доступны следующие опции

**flushAll** - true/false очистить все закеширование значения для каждого использованного инстанса memcached. По умолчанию равно true.

**maxConns** - число, максимальное количество неспользуемых сетевых соединений для каждого инстанса memcached. По умолчанию равно 20.

**dir** - директория из которой будут браться файлы для кеширования. По умолчанию текущая директория проекта.

**adid, dvid, gaid, idfa** - номера портов в которые сохранять соответствующие значения из файлов. Для данных значений должны быть подняты инстансы memcached.


Пример запуска с использованием опций:
```bash
go run ./memcLoader.go -adid=11212 -flushAll=false -dir=/tmp
```

Пример лога работы программы:
```linux
2022/05/05 21:50:30 Start caching files: 
20170929000010.tsv.gz 20170929000020.tsv.gz 20170929000030.tsv.gz 20170929000040.tsv.gz 20170929000050.tsv.gz 
2022/05/05 21:51:04 Prosessed file: 20170929000020.tsv.gz Bytes: 24111824 Chunks: 736 AllValues: 77005 Good Values: 77005 Err values: 0
2022/05/05 21:53:27 Prosessed file: 20170929000010.tsv.gz Bytes: 132739245 Chunks: 4051 AllValues: 422995 Good Values: 422995 Err values: 0
Prefixed current handling file 20170929000010.tsv.gz
Prefixed file 20170929000020.tsv.gz from buffer while goroutine working
2022/05/05 21:53:57 Prosessed file: 20170929000030.tsv.gz Bytes: 156967607 Chunks: 4791 AllValues: 500000 Good Values: 500000 Err values: 0
Prefixed current handling file 20170929000030.tsv.gz
2022/05/05 21:53:57 Prosessed file: 20170929000050.tsv.gz Bytes: 156910460 Chunks: 4789 AllValues: 500000 Good Values: 500000 Err values: 0
2022/05/05 21:56:04 8000 Chunks and 835363 values done for file 20170929000040.tsv.gz
2022/05/05 21:57:04 Prosessed file: 20170929000040.tsv.gz Bytes: 313852122 Chunks: 9579 AllValues: 1000000 Good Values: 1000000 Err values: 0
Prefixed current handling file 20170929000040.tsv.gz
Prefixed file 20170929000050.tsv.gz from buffer while goroutine working
```

Для проведения unit-тестов выполните:
```bash
go test
```
