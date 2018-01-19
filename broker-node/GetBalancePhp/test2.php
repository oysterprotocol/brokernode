<?php
    class CurlRequest
    {


        public function sendGet()
        {
            $ClientAddress = $_POST["Address"];
            $ContractAddress = $_POST["cAddress"];
        	$url = "https://api.etherscan.io/api?module=account&action=tokenbalance&contractaddress=".$ContractAddress."&address=".$ClientAddress."&tag=latest&apikey=R9DUZHGTW32FE3SK432G5NU2B1IR148RMS";
            //datos a enviar
            $data = array("a" => "a");
            //url contra la que atacamos
            $ch = curl_init($url);
            //a true, obtendremos una respuesta de la url, en otro caso, 
            //true si es correcto, false si no lo es
            curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
            //establecemos el verbo http que queremos utilizar para la petición
            curl_setopt($ch, CURLOPT_CUSTOMREQUEST, "GET");
            //enviamos el array data
            curl_setopt($ch, CURLOPT_POSTFIELDS,http_build_query($data));
            //obtenemos la respuesta
            $response = curl_exec($ch);
            // Se cierra el recurso CURL y se liberan los recursos del sistema
            curl_close($ch);
            if(!$response) {
                return false;
            }else{
            
                $object = json_decode($response);
                print "Your Balance is: ".$object->{'result'}; 
            }
        }
        }

     $new = new CurlRequest();

    $resultado = $new ->sendGet();
    
    
    //var_dump($resultado);

?>